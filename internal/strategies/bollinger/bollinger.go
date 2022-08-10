package bollinger

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies"
	indicators "tinkoff-invest-contest/internal/technical_indicators"
	"tinkoff-invest-contest/internal/utils"
)

type bollingerStrategy struct {
	indicator      *indicators.BollingerBands
	pointDeviation float64
}

type bollingerParams struct {
	Coef           float64 `json:"coef" yaml:"Coef"`
	PointDeviation float64 `json:"pointDev" yaml:"PointDeviation"`
}

func init() {
	strategies.Names = append(strategies.Names, "bollinger")
	strategies.JSONConstructors["bollinger"] = NewFromJSON
	strategies.DefaultsJSON["bollinger"] = GetDefaultsJSON
}

func NewFromJSON(s string) (strategies.Strategy, error) {
	p := bollingerParams{}

	err := json.Unmarshal([]byte(s), &p)
	if err != nil {
		return nil, err
	}

	return &bollingerStrategy{
		indicator:      indicators.NewBollingerBands(p.Coef),
		pointDeviation: p.PointDeviation,
	}, nil
}

func GetDefaultsJSON() string {
	defaults := bollingerParams{
		Coef:           3,
		PointDeviation: 0.0005,
	}
	bytes, err := json.MarshalIndent(&defaults, "", "  ")
	utils.MaybeCrash(err)
	return string(bytes)
}

func (b *bollingerStrategy) GetTradeSignal(marketData strategies.MarketData) (*utils.TradeSignal, map[string]any) {
	lowerBound, upperBound := b.indicator.Calculate(marketData.Candles)
	indicatorValues := map[string]any{
		"bollinger_lower_bound": lowerBound,
		"bollinger_upper_bound": upperBound,
	}

	currentCandle := marketData.Candles[len(marketData.Candles)-1]
	var signal *utils.TradeSignal
	if strategies.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), lowerBound, b.pointDeviation) &&
		utils.QuotationToFloat(currentCandle.Close) < ((lowerBound+upperBound)/2) {
		// Buy signal
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY}

	} else if strategies.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), upperBound, b.pointDeviation) {
		// Sell signal
		signal = &utils.TradeSignal{Direction: investapi.OrderDirection_ORDER_DIRECTION_SELL}
	}

	return signal, indicatorValues
}

func (*bollingerStrategy) GetOutputKeys() []string {
	return []string{
		"bollinger_lower_bound",
		"bollinger_upper_bound",
	}
}

func (b *bollingerStrategy) GetYAML() string {
	obj := bollingerParams{
		Coef:           b.indicator.GetCoef(),
		PointDeviation: b.pointDeviation,
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}

func (*bollingerStrategy) GetName() string {
	return "Bollinger Bands (R)"
}
