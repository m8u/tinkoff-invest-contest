package bots

import (
	"log"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/metrics"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorBot struct {
	figi           string
	candleInterval investapi.CandleInterval

	tradeEnv *tradeenv.TradeEnv
	charts   *metrics.Charts
}

func New(tradeEnv *tradeenv.TradeEnv, charts *metrics.Charts) *TechnicalIndicatorBot {
	bot := new(TechnicalIndicatorBot)
	bot.tradeEnv = tradeEnv
	bot.charts = charts

	bot.tradeEnv.InitChannels(bot.figi)

	return bot
}

func (bot *TechnicalIndicatorBot) loop() {
	// get the next candle from candle stream
	candle := <-bot.tradeEnv.Channels[bot.figi].Candle
	log.Println(candle.Close)
}

func (bot *TechnicalIndicatorBot) Serve() {
	for {
		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeInfo(bot.figi)
		utils.MaybeCrash(err)
		err = bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)

		bot.loop()
	}
}
