package strategy

import "tinkoff-invest-contest/internal/client/investapi"

type MarketData struct {
	Candles   []*investapi.HistoricCandle
	OrderBook *investapi.OrderBook
}
