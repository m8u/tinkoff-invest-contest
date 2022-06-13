package bots

import (
	"log"
	"tinkoff-invest-contest/internal/appstate"
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

func New(tradeEnv *tradeenv.TradeEnv, charts *metrics.Charts, figi string, candleInterval investapi.CandleInterval) *TechnicalIndicatorBot {
	bot := new(TechnicalIndicatorBot)

	bot.figi = figi
	bot.candleInterval = candleInterval

	bot.tradeEnv = tradeEnv
	bot.charts = charts

	bot.tradeEnv.InitChannels(bot.figi)

	return bot
}

func (bot *TechnicalIndicatorBot) loop() {
	for !appstate.ShouldExit {
		candle := <-bot.tradeEnv.Channels[bot.figi].Candle
		log.Println(candle)
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

	if bot.tradeEnv.IsSandbox {
		for _, account := range bot.tradeEnv.Accounts {
			_, err := bot.tradeEnv.Client.CloseSandboxAccount(account.Id)
			utils.MaybeCrash(err)
		}
	}
}
