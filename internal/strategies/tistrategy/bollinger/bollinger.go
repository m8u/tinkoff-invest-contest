package bollinger

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	indicators "tinkoff-invest-contest/internal/technical_indicators"
	"tinkoff-invest-contest/internal/utils"
)

type strategy struct {
	indicator      *indicators.BollingerBands
	pointDeviation float64
}

type params struct {
	Coef           float64 `json:"coef" yaml:"Coef"`
	PointDeviation float64 `json:"pointDev" yaml:"PointDeviation"`
}

func NewFromJSON(s string) (tistrategy.TechnicalIndicatorStrategy, error) {
	p := params{}

	err := json.Unmarshal([]byte(s), &p)
	if err != nil {
		return nil, err
	}

	return &strategy{
		indicator:      indicators.NewBollingerBands(p.Coef),
		pointDeviation: p.PointDeviation,
	}, nil
}

func GetDefaultsJSON() string {
	defaults := params{
		Coef:           3,
		PointDeviation: 0.001,
	}
	bytes, err := json.MarshalIndent(&defaults, "", "  ")
	utils.MaybeCrash(err)
	return string(bytes)
}

func (strategy *strategy) GetTradeSignal(candles []*investapi.HistoricCandle) (*utils.TradeSignal, map[string]any) {
	lowerBound, upperBound := strategy.indicator.Calculate(candles)
	indicatorValues := map[string]any{
		"bollinger_lower_bound": lowerBound,
		"bollinger_upper_bound": upperBound,
	}

	currentCandle := candles[len(candles)-1]
	var signal *utils.TradeSignal
	if tistrategy.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), lowerBound, strategy.pointDeviation) {
		// Сигнал к покупке
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY}

	} else if tistrategy.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), upperBound, strategy.pointDeviation) {
		// Сигнал к продаже
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_SELL}
	}

	return signal, indicatorValues
}

func (strategy *strategy) GetOutputKeys() []string {
	return []string{
		"bollinger_lower_bound",
		"bollinger_upper_bound",
	}
}

func (strategy *strategy) GetYAML() string {
	obj := params{
		Coef:           strategy.indicator.GetCoef(),
		PointDeviation: strategy.pointDeviation,
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}

func (strategy *strategy) GetName() string {
	return "Bollinger Bands (R)"
}
