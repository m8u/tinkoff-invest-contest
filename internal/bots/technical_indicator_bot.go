package bots

import (
	"log"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/dashboard"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/strategies/ti"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorBot struct {
	figi           string
	candleInterval investapi.CandleInterval
	window         int

	tradeEnv *tradeenv.TradeEnv

	strategy ti.TechnicalIndicatorStrategy
}

func New(tradeEnv *tradeenv.TradeEnv, figi string, candleInterval investapi.CandleInterval,
	window int, strategy ti.TechnicalIndicatorStrategy) *TechnicalIndicatorBot {
	bot := new(TechnicalIndicatorBot)

	bot.figi = figi
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

	return bot
}

func (bot *TechnicalIndicatorBot) loop() error {
	currentTimestamp := time.Time{}

	var candles []*investapi.HistoricCandle
	var err error

	for {
		select {
		// Get candle from stream
		case currentCandle := <-bot.tradeEnv.Channels[bot.figi].Candle:
			if currentCandle.Time.AsTime() != currentTimestamp {
				// On a new candle, get historic candles in amount of >= window
				candles, err = bot.tradeEnv.GetAtLeastNLastCandles(bot.figi, bot.candleInterval, bot.window)
				if err != nil {
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
			_, indicatorValues := bot.strategy.GetTradeSignal(
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
			break
		case <-appstate.ExitChan:
			return nil
		}
	}
}

func (bot *TechnicalIndicatorBot) Serve() {
	go func() {
		<-appstate.ExitChan
		bot.exitActions()
	}()

	for !appstate.ShouldExit {
		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)

		log.Printf("bot#%v is starting...", bot.figi)
		err = bot.loop()
		log.Printf("bot#%v has crashed with error: %v", bot.figi, err) // TODO: implement bot ids
	}
}

func (bot *TechnicalIndicatorBot) exitActions() {
	err := bot.tradeEnv.Client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	utils.MaybeCrash(err)
}
