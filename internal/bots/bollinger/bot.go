package bollinger

import (
	"log"
	"tinkoff-invest-contest/internal/metrics"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type Bot struct {
	figi     string
	tradeEnv *tradeenv.TradeEnv
	charts   *metrics.Charts
}

func New(tradeEnv *tradeenv.TradeEnv, charts *metrics.Charts) *Bot {
	bot := new(Bot)

	bot.tradeEnv = tradeEnv
	bot.charts = charts

	return bot
}

func (bot *Bot) loop() {
	// get the next candle from candle channel
	candle := (<-bot.tradeEnv.Instruments[bot.figi].CandleCh)
	log.Println(candle.Close)
}

func (bot *Bot) Serve() {
	for {
		utils.WaitForInternetConnection()
		bot.loop()
	}
}
