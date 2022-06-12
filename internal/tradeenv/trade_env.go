package tradeenv

import (
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/utils"
)

type TradingInstrument struct {
	candleInterval investapi.CandleInterval

	CandleCh      chan investapi.Candle
	OrderBookCh   chan investapi.OrderBook
	TradingStatus investapi.TradingStatus
}

type TradeEnv struct {
	token       string
	IsSandbox   bool
	Client      *client.Client
	Accounts    []*investapi.Account
	fee         float64
	Instruments map[string]TradingInstrument
}

func New(config config.Config) *TradeEnv {
	tradeEnv := new(TradeEnv)
	tradeEnv.Client = client.NewClient(config.Token)
	tradeEnv.Client.InitMarketDataStream()
	tradeEnv.IsSandbox = true
	tradeEnv.fee = config.Fee

	if config.IsSandbox {
		tradeEnv.Accounts = make([]*investapi.Account, 0)
		for i := 0; i < config.NumAccounts; i++ {
			accountResp, err := tradeEnv.Client.OpenSandboxAccount()
			utils.MaybeCrash(err)
			_, err = tradeEnv.Client.SandboxPayIn(accountResp.AccountId, "rub", config.Money)
			utils.MaybeCrash(err)

			tradeEnv.Accounts = append(tradeEnv.Accounts, &investapi.Account{Id: accountResp.AccountId})
		}
	} else {
		accounts, err := tradeEnv.Client.GetAccounts()
		utils.MaybeCrash(err)
		tradeEnv.Accounts = accounts
	}

	tradeEnv.Instruments = make(map[string]TradingInstrument)
	for _, instrument := range config.Instruments {
		err := tradeEnv.Client.SubscribeCandles(instrument.FIGI, investapi.SubscriptionInterval(instrument.CandleInterval))
		utils.MaybeCrash(err)
		err = tradeEnv.Client.SubscribeOrderBook(instrument.FIGI, instrument.OrderBookDepth)
		utils.MaybeCrash(err)
		err = tradeEnv.Client.SubscribeInfo(instrument.FIGI)
		utils.MaybeCrash(err)

		tradeEnv.Instruments[instrument.FIGI] = TradingInstrument{
			candleInterval: instrument.CandleInterval,
			CandleCh:       make(chan investapi.Candle),
			OrderBookCh:    make(chan investapi.OrderBook),
			TradingStatus:  investapi.TradingStatus{},
		}
	}

	tradeEnv.token = config.Token

	return tradeEnv
}
