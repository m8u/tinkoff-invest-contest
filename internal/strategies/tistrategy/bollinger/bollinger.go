package bollinger

import (
	"encoding/json"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	indicators "tinkoff-invest-contest/internal/technical_indicators"
	"tinkoff-invest-contest/internal/utils"
)

type strategy struct {
	indicator      *indicators.BollingerBands
	pointDeviation float64
}

func NewFromJsonString(s string) (tistrategy.TechnicalIndicatorStrategy, error) {
	params := struct {
		Coef           float64 `json:"coef"`
		PointDeviation float64 `json:"pointDev"`
	}{}

	err := json.Unmarshal([]byte(s), &params)
	if err != nil {
		return nil, err
	}

	return &strategy{
		indicator:      indicators.NewBollingerBands(params.Coef),
		pointDeviation: params.PointDeviation,
	}, nil
}

func (strategy *strategy) GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any) {
	lowerBound, upperBound := strategy.indicator.Calculate(candles)
	indicatorValues := map[string]any{
		"bollinger_lower_bound": lowerBound,
		"bollinger_upper_bound": upperBound,
	}

	currentCandle := candles[len(candles)-1]
	var signal *utils.TradeSignal
	if tistrategy.IsAroundPoint(utils.FloatFromQuotation(currentCandle.Close), lowerBound, strategy.pointDeviation) {
		// Сигнал к покупке
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY}

	} else if tistrategy.IsAroundPoint(utils.FloatFromQuotation(currentCandle.Close), upperBound, strategy.pointDeviation) {
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
