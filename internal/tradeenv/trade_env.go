package tradeenv

import (
	"log"
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

	Channels map[string]MarketDataChannelStack
}

func New(config config.Config) *TradeEnv {
	tradeEnv := new(TradeEnv)
	tradeEnv.Client = client.NewClient(config.Token)
	tradeEnv.Client.InitMarketDataStream()
	tradeEnv.IsSandbox = true
	tradeEnv.fee = config.Fee
	tradeEnv.Channels = make(map[string]MarketDataChannelStack)

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

	go tradeEnv.Client.RunMarketDataStreamLoop(tradeEnv.handleMarketDataStream)

	return tradeEnv
}

func (tradeEnv *TradeEnv) handleMarketDataStream(event *investapi.MarketDataResponse) {
	subscribeInfoResp := event.GetSubscribeInfoResponse()
	if subscribeInfoResp != nil {
		for _, s := range subscribeInfoResp.InfoSubscriptions {
			if s.SubscriptionStatus != investapi.SubscriptionStatus_SUBSCRIPTION_STATUS_SUCCESS {
				log.Fatalf("failed to subscribe to info (%v): %v", s.Figi, s.SubscriptionStatus.String())
			}
		}
	}
	subscribeCandlesResp := event.GetSubscribeCandlesResponse()
	if subscribeCandlesResp != nil {
		for _, s := range subscribeCandlesResp.CandlesSubscriptions {
			if s.SubscriptionStatus != investapi.SubscriptionStatus_SUBSCRIPTION_STATUS_SUCCESS {
				log.Fatalf("failed to subscribe to candles (%v): %v", s.Figi, s.SubscriptionStatus.String())
			}
		}
	}
	subscribeOrderBookResp := event.GetSubscribeOrderBookResponse()
	if subscribeOrderBookResp != nil {
		for _, s := range subscribeOrderBookResp.OrderBookSubscriptions {
			if s.SubscriptionStatus != investapi.SubscriptionStatus_SUBSCRIPTION_STATUS_SUCCESS {
				log.Fatalf("failed to subscribe to order book (%v): %v", s.Figi, s.SubscriptionStatus.String())
			}
		}
	}
	tradingStatus := event.GetTradingStatus()
	if tradingStatus != nil {
		tradeEnv.Channels[tradingStatus.Figi].TradingStatus <- tradingStatus
	}
	candle := event.GetCandle()
	if candle != nil {
		tradeEnv.Channels[tradingStatus.Figi].Candle <- candle
	}
	orderBook := event.GetOrderbook()
	if orderBook != nil {
		tradeEnv.Channels[tradingStatus.Figi].OrderBook <- orderBook
	}
}

type MarketDataChannelStack struct {
	TradingStatus chan *investapi.TradingStatus
	Candle        chan *investapi.Candle
	OrderBook     chan *investapi.OrderBook
}

func (tradeEnv TradeEnv) InitChannels(figi string) {
	tradeEnv.Channels[figi] = MarketDataChannelStack{
		TradingStatus: make(chan *investapi.TradingStatus),
		Candle:        make(chan *investapi.Candle),
		OrderBook:     make(chan *investapi.OrderBook),
	}
}
