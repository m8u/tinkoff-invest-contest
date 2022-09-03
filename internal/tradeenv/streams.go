package tradeenv

import (
	"log"
	"sync"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

var mu sync.Mutex

type subscriptions struct {
	candles   []*investapi.CandleInstrument
	info      []*investapi.InfoInstrument
	orderBook []*investapi.OrderBookInstrument
}

func (e *TradeEnv) SubscribeCandles(botId int, figi string, interval investapi.SubscriptionInterval) {
	err := e.Client.SubscribeCandles(figi, interval)
	utils.MaybeCrash(err)
	mu.Lock()
	if len(e.subscriptions.candles) < botId+1 {
		e.subscriptions.candles = append(e.subscriptions.candles,
			make([]*investapi.CandleInstrument, 1+botId-len(e.subscriptions.candles))...)
	}
	mu.Unlock()
	e.subscriptions.candles[botId] = &investapi.CandleInstrument{
		Figi:     figi,
		Interval: interval,
	}
}

func (e *TradeEnv) SubscribeInfo(botId int, figi string) {
	err := e.Client.SubscribeInfo(figi)
	utils.MaybeCrash(err)
	mu.Lock()
	if len(e.subscriptions.info) < botId+1 {
		e.subscriptions.info = append(e.subscriptions.info,
			make([]*investapi.InfoInstrument, 1+botId-len(e.subscriptions.info))...)
	}
	mu.Unlock()
	e.subscriptions.info[botId] = &investapi.InfoInstrument{
		Figi: figi,
	}
}

func (e *TradeEnv) SubscribeOrderBook(botId int, figi string, depth int32) {
	err := e.Client.SubscribeOrderBook(figi, depth)
	utils.MaybeCrash(err)
	mu.Lock()
	if len(e.subscriptions.orderBook) < botId+1 {
		e.subscriptions.orderBook = append(e.subscriptions.orderBook,
			make([]*investapi.OrderBookInstrument, 1+botId-len(e.subscriptions.orderBook))...)
	}
	mu.Unlock()
	e.subscriptions.orderBook[botId] = &investapi.OrderBookInstrument{
		Figi:  figi,
		Depth: depth,
	}
}

func (e *TradeEnv) UnsubscribeAll(botId int) {
	_ = e.Client.UnsubscribeCandles(e.subscriptions.candles[botId].Figi, e.subscriptions.candles[botId].Interval)
	_ = e.Client.UnsubscribeInfo(e.subscriptions.info[botId].Figi)
	_ = e.Client.UnsubscribeOrderBook(e.subscriptions.orderBook[botId].Figi, e.subscriptions.orderBook[botId].Depth)
	e.subscriptions.candles[botId] = nil
	e.subscriptions.info[botId] = nil
	e.subscriptions.orderBook[botId] = nil
}

func (e *TradeEnv) handleResubscribe() error {
	for _, subscription := range e.subscriptions.candles {
		if subscription == nil {
			continue
		}
		_ = e.Client.UnsubscribeCandles(
			subscription.Figi,
			subscription.Interval,
		)
		if err := e.Client.SubscribeCandles(
			subscription.Figi,
			subscription.Interval,
		); err != nil {
			return err
		}
	}
	for _, subscription := range e.subscriptions.info {
		if subscription == nil {
			continue
		}
		_ = e.Client.UnsubscribeInfo(
			subscription.Figi,
		)
		if err := e.Client.SubscribeInfo(
			subscription.Figi,
		); err != nil {
			return err
		}
	}
	for _, subscription := range e.subscriptions.orderBook {
		if subscription == nil {
			continue
		}
		_ = e.Client.UnsubscribeOrderBook(
			subscription.Figi,
			subscription.Depth,
		)
		if err := e.Client.SubscribeOrderBook(
			subscription.Figi,
			subscription.Depth,
		); err != nil {
			return err
		}
	}
	return nil
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
		for i, subscription := range e.subscriptions.info {
			if subscription == nil {
				continue
			}
			if subscription.Figi == tradingStatus.Figi {
				e.marketData[i].TradingStatus <- tradingStatus
			}
		}
	}
	candle := event.GetCandle()
	if candle != nil {
		for i, subscription := range e.subscriptions.candles {
			if subscription == nil {
				continue
			}
			if subscription.Figi == candle.Figi && subscription.Interval == candle.Interval {
				e.marketData[i].Candle <- candle
			}
		}
	}
	orderBook := event.GetOrderbook()
	if orderBook != nil {
		for i, subscription := range e.subscriptions.orderBook {
			if subscription == nil {
				continue
			}
			if subscription.Figi == orderBook.Figi && subscription.Depth == orderBook.Depth {
				e.marketData[i].OrderBook <- orderBook
			}
		}
	}
}

type MarketDataChannelStack struct {
	TradingStatus chan *investapi.TradingStatus
	Candle        chan *investapi.Candle
	OrderBook     chan *investapi.OrderBook
}

func (e *TradeEnv) InitNewMarketDataChannels() {
	e.marketData = append(e.marketData, &MarketDataChannelStack{
		TradingStatus: make(chan *investapi.TradingStatus, 1000),
		Candle:        make(chan *investapi.Candle, 1000),
		OrderBook:     make(chan *investapi.OrderBook, 1000),
	})
}

func (e *TradeEnv) GetMarketDataChannels(botId int) *MarketDataChannelStack {
	return e.marketData[botId]
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
