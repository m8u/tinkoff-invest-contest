package trade

import (
	"github.com/google/uuid"
	"log"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/grpc/tinkoff/investapi"
	"tinkoff-invest-contest/internal/metrics"
	"tinkoff-invest-contest/internal/strategy"
	"tinkoff-invest-contest/internal/utils"
)

type marketInfo struct {
	tradingStatus *investapi.TradingStatus
	currentCandle *investapi.Candle
}

type Bot struct {
	isSandbox      bool
	token          string
	client         *client.Client
	account        *investapi.Account
	figi           string
	candleInterval investapi.CandleInterval
	fee            float64
	strategyParams strategy.BollingerParams
	allowMargin    bool
	marketInfo     marketInfo
}

// NewSandboxBot создает нового бота для торговли в песочнице
func NewSandboxBot(token string, money float64, figi string, candleInterval investapi.CandleInterval,
	fee float64, strategyParams strategy.BollingerParams, allowMargin bool) *Bot {
	bot := new(Bot)
	var err error
	bot.client = client.NewClient(token)
	bot.isSandbox = true
	bot.figi = figi
	bot.candleInterval = candleInterval
	bot.strategyParams = strategyParams
	bot.fee = fee
	bot.allowMargin = allowMargin

	go bot.runStream()

	accountResp, err := bot.client.OpenSandboxAccount()
	utils.MaybeCrash(err)
	bot.account = &investapi.Account{Id: accountResp.AccountId}

	_, err = bot.client.SandboxPayIn(bot.account.Id, "rub", money)
	utils.MaybeCrash(err)

	bot.token = token

	return bot
}

// NewCombatBot создает нового бота для торговли на реальной бирже
func NewCombatBot(token string, figi string, candleInterval investapi.CandleInterval,
	strategyParams strategy.BollingerParams, allowMargin bool) *Bot {
	bot := new(Bot)
	var err error
	bot.client = client.NewClient(token)

	userInfo, err := bot.client.GetInfo()
	utils.MaybeCrash(err)

	bot.isSandbox = false
	bot.figi = figi
	bot.candleInterval = candleInterval
	bot.strategyParams = strategyParams
	bot.fee = utils.Fees[utils.Tariff(userInfo.Tariff)]
	bot.allowMargin = allowMargin

	go bot.runStream()

	accounts, err := bot.client.GetAccounts()
	utils.MaybeCrash(err)
	bot.account = accounts[0] // TODO: если несколько счетов, выбирается первый попавшийся

	bot.token = token

	return bot
}

// runStream запускает цикл чтения ивентов MarketDataStream, который самовосстанавливается в случае потери соединения
func (bot *Bot) runStream() {
	bot.client.InitMarketDataStream()
	for !appstate.ShouldExit {
		err := bot.client.SubscribeInfo(bot.figi)
		utils.MaybeCrash(err)
		if investapi.SubscriptionInterval(bot.candleInterval) >
			investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_FIVE_MINUTES {
			log.Fatalln("can't use this candle interval for realtime trading (yet)")
		}
		err = bot.client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)
		log.Println("market data stream is running")
		err = bot.client.RunMarketDataStreamLoop(func(resp *investapi.MarketDataResponse) {
			tradingStatus := resp.GetTradingStatus()
			if tradingStatus != nil {
				bot.marketInfo.tradingStatus = tradingStatus
			}
			currentCandle := resp.GetCandle()
			if currentCandle != nil {
				bot.marketInfo.currentCandle = currentCandle
			}
		})
		log.Println("market data stream has collapsed")

		time.Sleep(10 * time.Second)

		for appstate.NoInternetConnection {
			time.Sleep(time.Second)
		}

		log.Println("reopening market data stream...")

		bot.client.InitMarketDataStream()
	}
}

// Serve запускает основной цикл работы бота
func (bot *Bot) Serve(charts *metrics.Charts) {
	defer bot.client.UnsubscribeInfo(bot.figi)
	defer bot.client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	if bot.isSandbox {
		defer bot.client.CloseSandboxAccount(bot.account.Id)
	}

	var err error

	// получаем ранние свечи в количестве >= window
	for i := 0; len(*charts.Candles) < bot.strategyParams.Window; i++ {
		candles, err := client.GetCandlesForLastNDays(bot.client, bot.figi, i, bot.candleInterval)
		utils.MaybeCrash(err)
		*charts.Candles = append(candles, *charts.Candles...)
	}

	var positions *investapi.PositionsResponse
	if bot.isSandbox {
		positions, err = bot.client.GetSandboxPositions(bot.account.Id)
	} else {
		positions, err = bot.client.GetPositions(bot.account.Id)
	}
	utils.MaybeCrash(err)
	for _, money := range positions.Money {
		if money.Currency == "rub" {
			*charts.StartBalance = utils.FloatFromMoneyValue(money)
		}
	}

	// Дневной цикл
	for !appstate.ShouldExit {
		for bot.marketInfo.currentCandle == nil {
			// подождем пока не прилетит первая свеча из стрима
		}
		share, err := bot.client.ShareBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", bot.figi)
		utils.MaybeCrash(err)

		if bot.allowMargin &&
			((share.Dshort.Units == 0 && share.Dshort.Nano == 0) ||
				(share.Dlong.Units == 0 && share.Dlong.Nano == 0)) {
			log.Fatalf("can't margin-trade %v (%v)", share.Ticker, bot.figi)
		}

		if bot.isSandbox {
			positions, err = bot.client.GetSandboxPositions(bot.account.Id)
		} else {
			positions, err = bot.client.GetPositions(bot.account.Id)
		}
		utils.MaybeCrash(err)

		var moneyHave float64
		for _, money := range positions.Money {
			if money.Currency == "rub" {
				moneyHave = utils.FloatFromMoneyValue(money)
			}
		}
		var lotsHave int64
		if len(positions.Securities) > 0 {
			lotsHave = positions.Securities[0].Balance / int64(share.Lot)
		} else {
			lotsHave = 0
		}

		newDay := true
		// Свечной цикл
		// торгуем только в основной период
		for !appstate.ShouldExit && bot.marketInfo.tradingStatus.TradingStatus ==
			investapi.SecurityTradingStatus_SECURITY_TRADING_STATUS_NORMAL_TRADING {
			if newDay {
				log.Printf("NEW TRADING DAY STARTED\n"+
					"instrument: %v (%v)\n"+
					"bollinger window: %v\n"+
					"bollinger coef: %v\n"+
					"bollinger point dev.: %v\n"+
					"candle interval: %v\n"+
					"allow margin: %v\n"+
					"fee: %v\n"+
					"money have: %+v\n"+
					"lots have: %+v",
					share.Ticker,
					bot.figi,
					bot.strategyParams.Window,
					bot.strategyParams.BollingerCoef,
					bot.strategyParams.IntervalPointDeviation,
					utils.CandleIntervalsValuesToV1Names[bot.candleInterval],
					bot.allowMargin,
					bot.fee,
					moneyHave,
					lotsHave)
			}

			*charts.Candles = append(*charts.Candles, &investapi.HistoricCandle{
				Open:  bot.marketInfo.currentCandle.Open,
				High:  bot.marketInfo.currentCandle.High,
				Low:   bot.marketInfo.currentCandle.Low,
				Close: bot.marketInfo.currentCandle.Close,
				Time:  bot.marketInfo.currentCandle.Time,
			})
			currentCandleTS := bot.marketInfo.currentCandle.Time.AsTime()

			if len(*charts.Candles) > 1 {
				candles, err := bot.client.GetCandles(
					bot.figi,
					time.Now().Add(-3*utils.CandleIntervalsToDurations[bot.candleInterval]),
					time.Now(),
					bot.candleInterval,
				)
				utils.MaybeCrash(err)
				if len(candles) > 0 {
					(*charts.Candles)[len(*charts.Candles)-2] = candles[len(candles)-1]
				}
			}

			newCandle := true
			// Тиковый цикл (>=1 сек)
			for !appstate.ShouldExit && bot.marketInfo.currentCandle.Time.AsTime() == currentCandleTS {
				if newCandle {
					utils.WaitForInternetConnection()
				}

				// обновляем последнюю свечку на графике
				(*charts.Candles)[len(*charts.Candles)-1] = &investapi.HistoricCandle{
					Open:  bot.marketInfo.currentCandle.Open,
					High:  bot.marketInfo.currentCandle.High,
					Low:   bot.marketInfo.currentCandle.Low,
					Close: bot.marketInfo.currentCandle.Close,
					Time:  bot.marketInfo.currentCandle.Time,
				}

				tradeSignal := strategy.GetTradeSignal(
					bot.strategyParams,
					false,
					bot.marketInfo.currentCandle,
					newCandle,
					charts,
				)

				if tradeSignal != nil {
					if bot.isSandbox {
						positions, err = bot.client.GetSandboxPositions(bot.account.Id)
					} else {
						positions, err = bot.client.GetPositions(bot.account.Id)
					}
					utils.MaybeCrash(err)

					for _, money := range positions.Money {
						if money.Currency == "rub" {
							moneyHave = utils.FloatFromMoneyValue(money)
						}
					}
					if len(positions.Securities) > 0 {
						lotsHave = positions.Securities[0].Balance / int64(share.Lot)
					} else {
						lotsHave = 0
					}

					var marginAttributes *investapi.GetMarginAttributesResponse
					if !bot.isSandbox {
						marginAttributes, err = bot.client.GetMarginAttributes(bot.account.Id)
						utils.MaybeCrash(err)
					}

					var maxDealValue float64
					var lots int64
					var orderPrice float64

					switch tradeSignal.Direction {
					case investapi.OrderDirection_ORDER_DIRECTION_BUY:
						if bot.allowMargin {
							liquidPortfolio := utils.FloatFromMoneyValue(marginAttributes.LiquidPortfolio)
							startMargin := utils.FloatFromMoneyValue(marginAttributes.StartingMargin)
							maxDealValue = (liquidPortfolio - startMargin) / utils.FloatFromQuotation(share.Dlong)
						} else {
							maxDealValue = moneyHave
						}
						lots = int64(maxDealValue /
							(utils.FloatFromQuotation(bot.marketInfo.currentCandle.Close) *
								(1 + bot.fee) *
								float64(share.Lot)))
						if lots == 0 {
							goto NextTick
						}

						log.Printf("lots have: %+v; money have: %+v", lotsHave, moneyHave)

						orderPrice = utils.FloatFromQuotation(bot.marketInfo.currentCandle.Close)
						var order *investapi.PostOrderResponse
						if bot.isSandbox {
							order, err = bot.client.PostSandboxOrder(
								bot.figi,
								lots,
								0,
								investapi.OrderDirection_ORDER_DIRECTION_BUY,
								bot.account.Id,
								investapi.OrderType_ORDER_TYPE_MARKET,
								uuid.New().String())
						} else {
							order, err = bot.client.PostOrder(
								bot.figi,
								lots,
								0,
								investapi.OrderDirection_ORDER_DIRECTION_BUY,
								bot.account.Id,
								investapi.OrderType_ORDER_TYPE_MARKET,
								uuid.New().String())
						}
						if err != nil {
							log.Printf("!!! order error: %v", err)
						}
						if order.ExecutionReportStatus != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
							orderState := &investapi.OrderState{ExecutionReportStatus: investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW}
							for orderState.ExecutionReportStatus != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
								orderState, err = bot.client.GetOrderState(bot.account.Id, order.OrderId)
							}
							time.Sleep(5 * time.Second)
						}
						log.Printf("BUY %v for %v (executed: %v; status: %v)",
							order.LotsRequested, orderPrice,
							order.LotsExecuted, order.ExecutionReportStatus.String())

						break
					case investapi.OrderDirection_ORDER_DIRECTION_SELL:
						if bot.allowMargin && share.ShortEnabledFlag {
							liquidPortfolio := utils.FloatFromMoneyValue(marginAttributes.LiquidPortfolio)
							startMargin := utils.FloatFromMoneyValue(marginAttributes.StartingMargin)
							maxDealValue = (liquidPortfolio - startMargin) / utils.FloatFromQuotation(share.Dshort)
							lots = int64(maxDealValue /
								(utils.FloatFromQuotation(bot.marketInfo.currentCandle.Close) *
									(1 - bot.fee) *
									float64(share.Lot)))
						} else {
							lots = lotsHave
						}
						if lots == 0 {
							goto NextTick
						}

						log.Printf("lots have: %+v; money have: %+v", lotsHave, moneyHave)

						orderPrice = utils.FloatFromQuotation(bot.marketInfo.currentCandle.Close)
						var order *investapi.PostOrderResponse
						if bot.isSandbox {
							order, err = bot.client.PostSandboxOrder(
								bot.figi,
								lots,
								0,
								investapi.OrderDirection_ORDER_DIRECTION_SELL,
								bot.account.Id,
								investapi.OrderType_ORDER_TYPE_MARKET,
								uuid.New().String())
						} else {
							order, err = bot.client.PostOrder(
								bot.figi,
								lots,
								0,
								investapi.OrderDirection_ORDER_DIRECTION_SELL,
								bot.account.Id,
								investapi.OrderType_ORDER_TYPE_MARKET,
								uuid.New().String())
						}
						if err != nil {
							log.Printf("!!! order error: %v", err)
						}
						if order.ExecutionReportStatus != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
							orderState := &investapi.OrderState{ExecutionReportStatus: investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW}
							for orderState.ExecutionReportStatus != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
								orderState, err = bot.client.GetOrderState(bot.account.Id, order.OrderId)
							}
							time.Sleep(5 * time.Second)
						}
						log.Printf("SELL %v for %v (executed: %v; status: %v)",
							order.LotsRequested, orderPrice,
							order.LotsExecuted, order.ExecutionReportStatus.String())

						break
					}

					*charts.BalanceHistory = append(*charts.BalanceHistory, moneyHave)

					if len(*charts.Flags) > 0 {
						if (*charts.Flags)[len(*charts.Flags)-1][0].CandleIndex != len(*charts.Candles)-1 {
							*charts.Flags = append(*charts.Flags, make([]metrics.ChartsTradeFlag, 0))
						}
					} else {
						*charts.Flags = append(*charts.Flags, make([]metrics.ChartsTradeFlag, 0))
					}
					(*charts.Flags)[len(*charts.Flags)-1] = append((*charts.Flags)[len(*charts.Flags)-1],
						metrics.ChartsTradeFlag{
							tradeSignal.Direction,
							utils.FloatFromQuotation(bot.marketInfo.currentCandle.Close),
							lots * int64(share.Lot),
							len(*charts.Candles) - 1,
						},
					)
				}

			NextTick:
				newCandle = false
				time.Sleep(time.Second)
			}
			newDay = false
		}

		if !newDay && !appstate.ShouldExit { // !newDay - признак захода в свечной цикл
			log.Println("TRADING DAY HAS ENDED")
		}

		// чтобы при выходе функция завершилась и сработали defer-ы
		for i := 1; i <= 60 && !appstate.ShouldExit; i++ {
			time.Sleep(time.Second)
		}
	}
}
