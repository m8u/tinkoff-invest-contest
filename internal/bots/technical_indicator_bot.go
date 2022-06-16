package bots

import (
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/metrics"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorBot struct {
	figi           string
	candleInterval investapi.CandleInterval
	window         int

	tradeEnv *tradeenv.TradeEnv
	charts   *metrics.Charts
}

func New(tradeEnv *tradeenv.TradeEnv, charts *metrics.Charts, figi string, candleInterval investapi.CandleInterval,
	window int) *TechnicalIndicatorBot {
	bot := new(TechnicalIndicatorBot)

	bot.figi = figi
	bot.candleInterval = candleInterval
	bot.window = window

	bot.tradeEnv = tradeEnv
	bot.charts = charts

	bot.tradeEnv.InitChannels(bot.figi)

	err := db.CreateCandlesTable(bot.figi)
	utils.MaybeCrash(err)

	return bot
}

func (bot *TechnicalIndicatorBot) loop() {
	currentTimestamp := time.Time{}

	for !appstate.ShouldExit {
		// Get candle from stream
		candle := <-bot.tradeEnv.Channels[bot.figi].Candle
		if candle.Time.AsTime() != currentTimestamp {
			// On a new candle, update historic candles in amount of >= window
			candles, err := bot.tradeEnv.GetAtLeastNLastCandles(bot.figi, bot.candleInterval, bot.window)
			utils.MaybeCrash(err)
			err = db.InsertCandles(bot.figi, candles)
			utils.MaybeCrash(err)
			currentTimestamp = candle.Time.AsTime()
		}
	}
}

func (bot *TechnicalIndicatorBot) Serve() {
	for !appstate.ShouldExit {
		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)

		go func() {
			<-appstate.ExitChan
			bot.exitActions()
		}()

		bot.loop()
	}
}

func (bot *TechnicalIndicatorBot) exitActions() {
	err := bot.tradeEnv.Client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	utils.MaybeCrash(err)
}
