package tradeenv

import (
	"log"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type subscriptions struct {
	candles   map[string]investapi.CandleInstrument
	info      map[string]investapi.InfoInstrument
	orderBook map[string]investapi.OrderBookInstrument
}

func (e *TradeEnv) SubscribeCandles(botId string, figi string, interval investapi.SubscriptionInterval) {
	err := e.Client.SubscribeCandles(figi, interval)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.candles[botId] = investapi.CandleInstrument{
		Figi:     figi,
		Interval: interval,
	}
	e.mu.Unlock()
}

func (e *TradeEnv) SubscribeInfo(botId string, figi string) {
	err := e.Client.SubscribeInfo(figi)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.info[botId] = investapi.InfoInstrument{
		Figi: figi,
	}
	e.mu.Unlock()
}

func (e *TradeEnv) SubscribeOrderBook(botId string, figi string, depth int32) {
	err := e.Client.SubscribeOrderBook(figi, depth)
	utils.MaybeCrash(err)

	e.mu.Lock()
	e.subscriptions.orderBook[botId] = investapi.OrderBookInstrument{
		Figi:  figi,
		Depth: depth,
	}
	e.mu.Unlock()
}

func (e *TradeEnv) UnsubscribeAll(botId string) {
	e.mu.Lock()
	_ = e.Client.UnsubscribeCandles(e.subscriptions.candles[botId].Figi, e.subscriptions.candles[botId].Interval)
	_ = e.Client.UnsubscribeInfo(e.subscriptions.info[botId].Figi)
	_ = e.Client.UnsubscribeOrderBook(e.subscriptions.orderBook[botId].Figi, e.subscriptions.orderBook[botId].Depth)
	delete(e.subscriptions.candles, botId)
	delete(e.subscriptions.info, botId)
	delete(e.subscriptions.orderBook, botId)
	e.mu.Unlock()
}

func (e *TradeEnv) handleResubscribe() {
	e.mu.RLock()
	for _, subscription := range e.subscriptions.candles {
		err := e.Client.SubscribeCandles(
			subscription.Figi,
			subscription.Interval,
		)
		utils.MaybeCrash(err)
	}
	for _, subscription := range e.subscriptions.info {
		err := e.Client.SubscribeInfo(
			subscription.Figi,
		)
		utils.MaybeCrash(err)
	}
	for _, subscription := range e.subscriptions.orderBook {
		err := e.Client.SubscribeOrderBook(
			subscription.Figi,
			subscription.Depth,
		)
		utils.MaybeCrash(err)
	}
	e.mu.RUnlock()
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
