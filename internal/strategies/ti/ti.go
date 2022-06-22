package ti

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorStrategy interface {
	GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any)
	GetDescriptor() []string
}
