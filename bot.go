package main

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"reflect"
	"strconv"
	"time"
)

type marketInfo struct {
	instrumentInfo sdk.InstrumentInfo
	currentCandle  sdk.Candle
}

type Bot struct {
	isSandbox       bool
	token           string
	restClient      *RestClientV2
	streamingClient *sdk.StreamingClient
	account         sdk.Account
	figi            string
	candleInterval  sdk.CandleInterval
	fee             float64
	strategyParams  StrategyParams
	allowMargin     bool
	marketInfo      marketInfo
}

// NewSandboxBot создает нового бота для торговли в песочнице
func NewSandboxBot(token string, money float64, figi string, candleInterval sdk.CandleInterval,
	fee float64, strategyParams StrategyParams, allowMargin bool) *Bot {
	bot := new(Bot)
	var err error
	bot.restClient = &RestClientV2{token: token, appname: AppName}

	bot.isSandbox = true
	bot.figi = figi
	bot.candleInterval = candleInterval
	bot.strategyParams = strategyParams
	bot.fee = fee
	bot.allowMargin = allowMargin

	bot.streamingClient, err = sdk.NewStreamingClient(nil, token)
	MaybeCrash(err)
	go bot.runMarketDataStream()

	bot.account, err = bot.restClient.OpenSandboxAccount()
	MaybeCrash(err)

	_, err = bot.restClient.SandboxPayIn(bot.account.ID, sdk.RUB, money)
	MaybeCrash(err)

	bot.token = token

	return bot
}

// NewCombatBot создает нового бота для торговли на реальной бирже
func NewCombatBot(token string, figi string, candleInterval sdk.CandleInterval,
	strategyParams StrategyParams, allowMargin bool) *Bot {
	bot := new(Bot)
	var err error
	bot.restClient = &RestClientV2{token: token, appname: AppName}

	userInfo, err := bot.restClient.GetInfo()
	MaybeCrash(err)

	bot.isSandbox = false
	bot.figi = figi
	bot.candleInterval = candleInterval
	bot.strategyParams = strategyParams
	bot.fee = Fees[userInfo.Tariff]
	bot.allowMargin = allowMargin

	bot.streamingClient, err = sdk.NewStreamingClient(nil, token)
	MaybeCrash(err)
	go bot.runMarketDataStream()

	accounts, err := bot.restClient.GetAccounts()
	MaybeCrash(err)
	bot.account = accounts[0] // TODO: если несолько счетов, выбирается первый попавшийся

	bot.token = token

	return bot
}

// runMarketDataStream запускает цикл чтения ивентов StreamingAPI, который самовосстанавливается в случае потери соединения
func (bot *Bot) runMarketDataStream() {
	for !ShouldExit {
		// подписываемся на InstrumentInfo чтобы получать статус торговли
		err := bot.streamingClient.SubscribeInstrumentInfo(bot.figi, "0")
		MaybeCrash(err)
		// и на Candle чтобы получать текущую цену
		err = bot.streamingClient.SubscribeCandle(bot.figi, bot.candleInterval, "1")
		MaybeCrash(err)
		err = bot.streamingClient.RunReadLoop(func(event any) error {
			eventValue := reflect.ValueOf(event)
			switch eventValue.FieldByName("FullEvent").FieldByName("Name").String() {
			case "instrument_info":
				bot.marketInfo.instrumentInfo = eventValue.Interface().(sdk.InstrumentInfoEvent).Info
				break
			case "candle":
				bot.marketInfo.currentCandle = eventValue.Interface().(sdk.CandleEvent).Candle
				break
			case "error":
				log.Fatalln(eventValue)
			}
			return nil
		})
		log.Println(err.Error())

		time.Sleep(10 * time.Second)
		log.Println("reopening market data stream...")

		err = bot.streamingClient.Close()
		if err != nil {
			log.Println(err)
		}
		bot.streamingClient, err = sdk.NewStreamingClient(nil, bot.token)
		if err != nil {
			log.Println(err)
		}
	}
}

// Serve запускает основной цикл работы бота
func (bot *Bot) Serve(charts *Charts) {
	defer bot.streamingClient.UnsubscribeInstrumentInfo(bot.figi, "0")
	defer bot.streamingClient.UnsubscribeCandle(bot.figi, bot.candleInterval, "1")
	if bot.isSandbox {
		defer bot.restClient.CloseSandboxAccount(bot.account.ID)
	}

	// получаем ранние свечи в количестве >= window
	for i := 0; len(charts.Candles) < bot.strategyParams.Window; i++ {
		candles, err := GetCandlesForLastNDays(bot.restClient, bot.figi, i, bot.candleInterval)
		MaybeCrash(err)
		charts.Candles = append(candles, charts.Candles...)
	}

	// Дневной цикл
	for !ShouldExit { // TODO: интернет соединение как-то восстанавливать
		share, err := bot.restClient.ShareBy(FIGI, "", bot.figi)
		MaybeCrash(err)

		if bot.allowMargin &&
			((share.Instrument.Dshort.Units == "0" && share.Instrument.Dshort.Nano == 0) ||
				(share.Instrument.Dlong.Units == "0" && share.Instrument.Dlong.Nano == 0)) {
			log.Fatalf("can't margin-trade %v (%v)", share.Instrument.Ticker, bot.figi)
		}

		portfolio, err := bot.restClient.GetPortfolio(bot.account.ID, bot.isSandbox)
		MaybeCrash(err)

		moneyHave := portfolio.Currencies[0].Balance
		var lotsHave int
		if len(portfolio.Positions) > 0 {
			lotsHave = int(portfolio.Positions[0].Balance) / share.Instrument.Lot
		} else {
			lotsHave = 0
		}

		newDay := true
		// Свечной цикл
		// торгуем только в основной период
		for !ShouldExit && bot.marketInfo.instrumentInfo.TradeStatus == sdk.NormalTrading {
			if newDay {
				log.Printf("\n"+
					"NEW TRADING DAY STARTED\n"+
					"instrument: %v (%v)\n"+
					"allow margin: %t\n"+
					"money have: %+v\n"+
					"lots have: %+v",
					share.Instrument.Ticker, bot.figi, bot.allowMargin, moneyHave, lotsHave)
			}

			charts.Candles = append(charts.Candles, bot.marketInfo.currentCandle)
			currentCandleTS := bot.marketInfo.currentCandle.TS

			if len(charts.Candles) > 1 {
				candles, err := bot.restClient.GetCandles(
					time.Now().Add(-3*CandleIntervalsToDurations[bot.candleInterval]),
					time.Now(),
					bot.candleInterval,
					bot.figi,
				)
				MaybeCrash(err)
				if len(candles) > 0 {
					charts.Candles[len(charts.Candles)-2] = candles[len(candles)-1]
				}
			}

			newCandle := true
			// Тиковый цикл (>=1 сек)
			for !ShouldExit && bot.marketInfo.currentCandle.TS == currentCandleTS {
				charts.Candles[len(charts.Candles)-1] = bot.marketInfo.currentCandle

				tradeSignal := GetTradeSignal(
					bot.strategyParams,
					false,
					bot.marketInfo.currentCandle,
					newCandle,
					charts,
				)

				if tradeSignal != nil {
					portfolio, err := bot.restClient.GetPortfolio(bot.account.ID, bot.isSandbox)
					MaybeCrash(err)

					moneyHave = portfolio.Currencies[0].Balance
					if len(portfolio.Positions) > 0 {
						lotsHave = int(portfolio.Positions[0].Balance) / share.Instrument.Lot
					} else {
						lotsHave = 0
					}
					charts.BalanceHistory = append(charts.BalanceHistory, moneyHave)
					
					var marginAttributes MarginAttributes
					if bot.isSandbox {
					    marginAttributes, err = bot.restClient.GetMarginAttributes(bot.account.ID)
					    MaybeCrash(err)
					}

					var maxDealValue float64
					var lots int

					switch tradeSignal.Direction {
					case sdk.BUY:
						if bot.allowMargin {
							liquidPortfolio, _ := strconv.Atoi(marginAttributes.LiquidPortfolio.Units)
							startMargin, _ := strconv.Atoi(marginAttributes.StartingMargin.Units)
							maxDealValue = (float64(liquidPortfolio) - float64(startMargin)) / share.Instrument.Dlong.ToFloat()
						} else {
							maxDealValue = moneyHave
						}
						lots = int(maxDealValue / (bot.marketInfo.currentCandle.ClosePrice * (1 + bot.fee) * float64(share.Instrument.Lot)))
						if lots == 0 {
							goto NextTick
						}

						log.Printf("lots have: %+v; money have: %+v", lotsHave, moneyHave)
						
						order, err := bot.restClient.PostMarketOrder(bot.figi, lots, sdk.BUY, bot.account.ID, bot.isSandbox)
						if err != nil {
							log.Println("!!! order error: " + err.Error())
						}

						log.Printf("BUY %v for %v (executed: %v; status: %v)",
							order.RequestedLots, bot.marketInfo.currentCandle.ClosePrice,
							order.ExecutedLots, order.Status)
						break
					case sdk.SELL:
						if bot.allowMargin && share.Instrument.ShortEnabledFlag {
							liquidPortfolio, _ := strconv.Atoi(marginAttributes.LiquidPortfolio.Units)
							startMargin, _ := strconv.Atoi(marginAttributes.StartingMargin.Units)
							maxDealValue = (float64(liquidPortfolio) - float64(startMargin)) / share.Instrument.Dshort.ToFloat()
							lots = int(maxDealValue / (bot.marketInfo.currentCandle.ClosePrice * (1 - bot.fee) * float64(share.Instrument.Lot)))
						} else {
							lots = lotsHave
						}
						if lots == 0 {
							goto NextTick
						}

						log.Printf("lots have: %+v; money have: %+v", lotsHave, moneyHave)

						order, err := bot.restClient.PostMarketOrder(bot.figi, lots, sdk.SELL, bot.account.ID, bot.isSandbox)
						if err != nil {
							log.Println("!!! order error: " + err.Error())
						}

						log.Printf("SELL %v for %v (executed: %v; status: %v)",
							order.RequestedLots, bot.marketInfo.currentCandle.ClosePrice,
							order.ExecutedLots, order.Status)
						break
					}
					if len(charts.Flags) == 0 || charts.Flags[len(charts.Flags)-1][0].CandleIndex != len(charts.Candles)-1 {
						charts.Flags = append(charts.Flags, make([]ChartsTradeFlag, 0))
					} else {
						charts.Flags[len(charts.Flags)-1] = append(charts.Flags[len(charts.Flags)-1],
							ChartsTradeFlag{
								tradeSignal.Direction,
								bot.marketInfo.currentCandle.ClosePrice,
								lots * share.Instrument.Lot,
								len(charts.Candles) - 1,
							},
						)
					}
				}

			NextTick:
				newCandle = false
				time.Sleep(time.Second)
			}
			newDay = false
		}

		if !newDay { // признак захода в свечной цикл
			log.Println("TRADING DAY HAS ENDED")
		}

		// чтобы при выходе функция завершилась и сработали defer-ы
		for i := 1; i <= 60 && !ShouldExit; i++ {
			time.Sleep(time.Second)
		}
	}
}
