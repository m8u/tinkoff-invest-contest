package tradeenv

import (
	"tinkoff-invest-contest/internal/client"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/utils"
)

type TradeEnv struct {
	token     string
	IsSandbox bool
	Client    *client.Client
	Accounts  []*investapi.Account
	fee       float64

	Channels map[string]MarketDataChannels
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

	tradeEnv.token = config.Token

	return tradeEnv
}

type MarketDataChannels struct {
	TradingStatus chan investapi.TradingStatus
	Candle        chan investapi.Candle
	OrderBook     chan investapi.OrderBook
}

func (tradeEnv TradeEnv) InitChannels(figi string) {
	tradeEnv.Channels[figi] = MarketDataChannels{
		TradingStatus: make(chan investapi.TradingStatus),
		Candle:        make(chan investapi.Candle),
		OrderBook:     make(chan investapi.OrderBook),
	}
}
