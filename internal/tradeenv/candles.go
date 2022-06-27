package tradeenv

import (
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
)

func (e *TradeEnv) GetCandlesFor1NthDayBeforeNow(figi string,
	candleInterval investapi.CandleInterval, n int) ([]*investapi.HistoricCandle, error) {
	candles, err := e.Client.GetCandles(
		figi,
		time.Now().Add(-time.Duration(n+1)*24*time.Hour),
		time.Now().Add(-time.Duration(n)*24*time.Hour),
		candleInterval,
	)
	if err != nil {
		return nil, err
	}
	return candles, nil
}

func (e *TradeEnv) GetAtLeastNLastCandles(figi string,
	candleInterval investapi.CandleInterval, n int) ([]*investapi.HistoricCandle, error) {
	candles := make([]*investapi.HistoricCandle, 0)
	for i := 0; len(candles) < n; i++ {
		portion, err := e.GetCandlesFor1NthDayBeforeNow(figi, candleInterval, i)
		if err != nil {
			return nil, err
		}
		candles = append(portion, candles...)
	}
	return candles, nil
}
