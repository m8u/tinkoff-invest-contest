package tistrategy

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorStrategy interface {
	GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any)
	GetOutputKeys() []string
}

var JSONConstructors map[string]func(string) (TechnicalIndicatorStrategy, error)
var DefaultsJSON map[string]func() string
