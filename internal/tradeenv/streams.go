package tradeenv

import (
	"log"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type subscriptions struct {
	candles   []investapi.CandleInstrument
	info      []investapi.InfoInstrument
	orderBook []investapi.OrderBookInstrument
}

func (e *TradeEnv) SubscribeCandles(figi string, interval investapi.SubscriptionInterval) {
	err := e.Client.SubscribeCandles(figi, interval)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.candles = append(e.subscriptions.candles, investapi.CandleInstrument{
		Figi:     figi,
		Interval: interval,
	})
	e.mu.Unlock()
}

func (e *TradeEnv) SubscribeInfo(figi string) {
	err := e.Client.SubscribeInfo(figi)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.info = append(e.subscriptions.info, investapi.InfoInstrument{
		Figi: figi,
	})
	e.mu.Unlock()
}

func (e *TradeEnv) SubscribeOrderBook(figi string, depth int32) {
	err := e.Client.SubscribeOrderBook(figi, depth)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.orderBook = append(e.subscriptions.orderBook, investapi.OrderBookInstrument{
		Figi:  figi,
		Depth: depth,
	})
	e.mu.Unlock()
}

func (e *TradeEnv) handleResubscribe() {
	for i := 0; i < len(e.subscriptions.candles); i++ {
		err := e.Client.SubscribeCandles(
			e.subscriptions.candles[i].Figi,
			e.subscriptions.candles[i].Interval,
		)
		utils.MaybeCrash(err)
	}
	for i := 0; i < len(e.subscriptions.info); i++ {
		err := e.Client.SubscribeInfo(
			e.subscriptions.info[i].Figi,
		)
		utils.MaybeCrash(err)
	}
	for i := 0; i < len(e.subscriptions.orderBook); i++ {
		err := e.Client.SubscribeOrderBook(
			e.subscriptions.orderBook[i].Figi,
			e.subscriptions.orderBook[i].Depth,
		)
		utils.MaybeCrash(err)
	}
}

func (e *TradeEnv) handleMarketDataStream(event *investapi.MarketDataResponse) {
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
		e.marketData[tradingStatus.Figi].TradingStatus <- tradingStatus
	}
	candle := event.GetCandle()
	if candle != nil {
		e.marketData[candle.Figi].Candle <- candle
	}
	orderBook := event.GetOrderbook()
	if orderBook != nil {
		e.marketData[orderBook.Figi].OrderBook <- orderBook
	}
}

type MarketDataChannelStack struct {
	TradingStatus chan *investapi.TradingStatus
	Candle        chan *investapi.Candle
	OrderBook     chan *investapi.OrderBook
}

func (e *TradeEnv) InitMarketDataChannels(figi string) {
	e.mu.Lock()
	e.marketData[figi] = &MarketDataChannelStack{
		TradingStatus: make(chan *investapi.TradingStatus),
		Candle:        make(chan *investapi.Candle, 100),
		OrderBook:     make(chan *investapi.OrderBook, 100),
	}
	e.mu.Unlock()
}

func (e *TradeEnv) GetMarketDataChannels(figi string) *MarketDataChannelStack {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.marketData[figi]
}

func (e *TradeEnv) InitTradesChannels(accountIds []string) {
	e.Client.InitTradesStream(accountIds)
	e.mu.Lock()
	for _, id := range accountIds {
		e.trades[id] = make(chan *investapi.OrderTrades)
	}
	e.mu.Unlock()
}

func (e *TradeEnv) handleTradesStream(resp *investapi.TradesStreamResponse) {
	orderTrades := resp.GetOrderTrades()
	if orderTrades == nil {
		return
	}
	e.trades[orderTrades.AccountId] <- orderTrades
}

func (e *TradeEnv) GetTradesChannel(accountId string) chan *investapi.OrderTrades {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.trades[accountId]
}
