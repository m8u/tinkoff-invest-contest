package bots

import (
	"log"
	"math"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/dashboard"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorBot struct {
	figi           string
	instrumentType utils.InstrumentType
	candleInterval investapi.CandleInterval
	window         int
	allowMargin    bool

	tradeEnv          *tradeenv.TradeEnv
	occupiedAccountId string

	strategy tistrategy.TechnicalIndicatorStrategy
}

func New(tradeEnv *tradeenv.TradeEnv, figi string, instrumentType utils.InstrumentType,
	candleInterval investapi.CandleInterval, window int, strategy tistrategy.TechnicalIndicatorStrategy) *TechnicalIndicatorBot {
	bot := new(TechnicalIndicatorBot)

	bot.figi = figi
	bot.instrumentType = instrumentType
	bot.candleInterval = candleInterval
	bot.window = window
	bot.strategy = strategy

	bot.tradeEnv = tradeEnv

	bot.tradeEnv.InitChannels(bot.figi)

	err := db.CreateCandlesTable(bot.figi)
	utils.MaybeCrash(err)

	db.CreateIndicatorValuesTable(bot.figi, strategy.GetDescriptor())

	err = dashboard.AddBotDashboard(bot.figi)
	utils.MaybeCrash(err)

	appstate.PostExitActionsWG.Add(1)

	return bot
}

func (bot *TechnicalIndicatorBot) loop() error {
	currentTimestamp := time.Time{}

	var candles []*investapi.HistoricCandle
	var err error

	instrument, err := bot.tradeEnv.Client.InstrumentByFigi(bot.figi, bot.instrumentType)
	if err != nil {
		log.Println(utils.PrettifyError(err))
		return err
	}

	var prevDirection investapi.OrderDirection
	for !appstate.ShouldExit {
		select {
		// Get candle from stream
		case currentCandle := <-bot.tradeEnv.Channels[bot.figi].Candle:
			if currentCandle.Time.AsTime() != currentTimestamp {
				// On a new candle, get historic candles in amount of >= window
				candles, err = bot.tradeEnv.GetAtLeastNLastCandles(bot.figi, bot.candleInterval, bot.window)
				if err != nil {
					log.Println(utils.PrettifyError(err))
					return err
				}
				// Trim excessive candles
				candles = candles[len(candles)-(bot.window-1):]
				go func() {
					db.InsertCandles(bot.figi, candles)
				}()
				currentTimestamp = currentCandle.Time.AsTime()
			}
			go func() {
				db.UpdateLastCandle(bot.figi, currentCandle)
			}()

			// Get trade signal
			signal, indicatorValues := bot.strategy.GetTradeSignal(
				append(candles,
					&investapi.HistoricCandle{
						Open:   currentCandle.Open,
						High:   currentCandle.High,
						Low:    currentCandle.Low,
						Close:  currentCandle.Close,
						Volume: currentCandle.Volume,
					},
				),
			)
			indicatorValues["time"] = currentCandle.Time.AsTime()
			go func() {
				db.AddIndicatorValues(bot.figi, indicatorValues)
			}()

			if signal != nil {
				// Get unoccupied account or use the existing one,
				// and determine lot quantity for the deal (either buy or sell)
				var lots int64
				shouldUnoccupyAccount := false
				if bot.occupiedAccountId == "" {
					accountId, unlock := bot.tradeEnv.GetUnoccupiedAccount()
					if accountId == "" {
						unlock()
						continue
					}
					maxDealValue, err := bot.tradeEnv.CalculateMaxDealValue(
						accountId,
						signal.Direction,
						instrument,
						currentCandle.Close,
						bot.allowMargin,
					)
					if err != nil {
						log.Println(utils.PrettifyError(err))
						unlock()
						return err
					}
					lots = bot.tradeEnv.CalculateLotsCanAfford(signal.Direction, maxDealValue, instrument, currentCandle.Close)
					if lots == 0 {
						unlock()
						continue
					}
					bot.tradeEnv.SetAccountOccupied(accountId)
					unlock()
					bot.occupiedAccountId = accountId
				} else if signal.Direction != prevDirection {
					shouldUnoccupyAccount = true
					lots, err = bot.tradeEnv.GetLotsHave(bot.occupiedAccountId, instrument)
					if err != nil {
						log.Println(utils.PrettifyError(err))
						return err
					}
					lots = int64(math.Abs(float64(lots)))
				} else {
					continue
				}

				// Place an order and wait for it to be filled
				log.Printf("\n[Order]\n%v lots for %v\n%v\nfigi: %v\naccount: %v\n",
					lots,
					utils.FloatFromQuotation(currentCandle.Close),
					signal.Direction.String(),
					bot.figi,
					bot.occupiedAccountId,
				)
				order, err := bot.tradeEnv.DoOrder(
					bot.figi,
					lots,
					utils.FloatFromQuotation(currentCandle.Close),
					signal.Direction,
					bot.occupiedAccountId,
					investapi.OrderType_ORDER_TYPE_MARKET,
				)
				if err != nil {
					log.Printf("order error: %v, message: %v", utils.PrettifyError(err), order.Message)
					return err
				}

				if shouldUnoccupyAccount {
					bot.tradeEnv.SetAccountUnoccupied(bot.occupiedAccountId)
					bot.occupiedAccountId = ""
				}

				prevDirection = signal.Direction
			}

		default:
			// Don't block
		}
	}
	return nil
}

func (bot *TechnicalIndicatorBot) Serve() {
	go func() {
		appstate.ExitActionsWG.Wait()
		bot.exitActions()
	}()

	for !appstate.ShouldExit {
		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)

		log.Printf("bot#%v is starting...", bot.figi)
		err = bot.loop()
		if err != nil {
			log.Printf("bot#%v has crashed, restarting...", bot.figi) // TODO: implement bot ids
			time.Sleep(10 * time.Second)
		}
	}
}

func (bot *TechnicalIndicatorBot) exitActions() {
	defer appstate.PostExitActionsWG.Done()

	err := bot.tradeEnv.Client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	if err != nil {
		log.Println(utils.PrettifyError(err))
	}
}
