package client

import (
	"time"
	"tinkoff-invest-contest/internal/grpc/tinkoff/investapi"
)

// GetCandlesForLastNDays загружает свечи за заданное кол-во последних дней
func GetCandlesForLastNDays(client *Client, figi string, n int,
	candleInterval investapi.CandleInterval) ([]*investapi.HistoricCandle, error) {
	candles := make([]*investapi.HistoricCandle, 0)
	for day := n; day >= 0; day-- {
		portion, err := client.GetCandles(
			figi,
			time.Now().AddDate(0, 0, -day-1),
			time.Now().AddDate(0, 0, -day),
			candleInterval,
		)
		if err != nil {
			return nil, err
		}
		for _, currentCandle := range portion {
			candles = append(candles, currentCandle)
		}
	}
	return candles, nil
}
