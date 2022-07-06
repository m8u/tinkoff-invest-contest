package tistrategy

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorStrategy interface {
	GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any)
	GetDescriptor() []string
}

var JsonConstructors map[string]func(string) (TechnicalIndicatorStrategy, error)
