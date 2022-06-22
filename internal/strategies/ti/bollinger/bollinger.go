package bollinger

import (
	"tinkoff-invest-contest/internal/client/investapi"
	indicators "tinkoff-invest-contest/internal/technical_indicators"
	"tinkoff-invest-contest/internal/utils"
)

type strategy struct {
	indicator      *indicators.BollingerBands
	pointDeviation float64
}

func New(bollingerCoef float64, pointDeviation float64) *strategy {
	return &strategy{
		indicator:      indicators.NewBollingerBands(bollingerCoef),
		pointDeviation: pointDeviation,
	}
}

func (strategy *strategy) GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any) {
	lowerBound, upperBound := strategy.indicator.Calculate(candles)
	indicatorValues := map[string]any{
		"bollinger_lower_bound": lowerBound,
		"bollinger_upper_bound": upperBound,
	}

	currentCandle := candles[len(candles)-1]
	var signal *utils.TradeSignal
	if utils.IsAroundPoint(utils.FloatFromQuotation(currentCandle.Close), lowerBound, strategy.pointDeviation) {
		// Сигнал к покупке
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY}

	} else if utils.IsAroundPoint(utils.FloatFromQuotation(currentCandle.Close), upperBound, strategy.pointDeviation) {
		// Сигнал к продаже
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_SELL}
	}

	return signal, indicatorValues
}

func (strategy *strategy) GetDescriptor() []string {
	return []string{
		"bollinger_lower_bound",
		"bollinger_upper_bound",
	}
}
