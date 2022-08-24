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
	strategyName := "bollinger"
	strategies.Names = append(strategies.Names, strategyName)
	strategies.JSONConstructors[strategyName] = NewFromJSON
	strategies.DefaultsJSON[strategyName] = GetDefaultsJSON
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

func (s *bollingerStrategy) GetTradeSignal(instrument utils.InstrumentInterface, marketData strategies.MarketData) (*strategies.TradeSignal, map[string]any) {
	lowerBound, upperBound := s.indicator.Calculate(marketData.Candles)
	indicatorValues := map[string]any{
		"bollinger_lower_bound": lowerBound,
		"bollinger_upper_bound": upperBound,
	}

	currentCandle := marketData.Candles[len(marketData.Candles)-1]
	var signal *strategies.TradeSignal
	if strategies.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), lowerBound, s.pointDeviation) &&
		utils.QuotationToFloat(currentCandle.Close) < ((lowerBound+upperBound)/2) {
		// Buy signal
		signal = strategies.NewTradeSignal(
			investapi.OrderDirection_ORDER_DIRECTION_BUY,
			investapi.OrderType_ORDER_TYPE_MARKET,
			currentCandle.Close,
		)

	} else if strategies.IsAroundPoint(utils.QuotationToFloat(currentCandle.Close), upperBound, s.pointDeviation) {
		// Sell signal
		signal = strategies.NewTradeSignal(
			investapi.OrderDirection_ORDER_DIRECTION_SELL,
			investapi.OrderType_ORDER_TYPE_MARKET,
			currentCandle.Close,
		)
	}

	return signal, indicatorValues
}

func (*bollingerStrategy) GetOutputKeys() []string {
	return []string{
		"bollinger_lower_bound",
		"bollinger_upper_bound",
	}
}

func (s *bollingerStrategy) GetYAML() string {
	obj := bollingerParams{
		Coef:           s.indicator.GetCoef(),
		PointDeviation: s.pointDeviation,
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}

func (*bollingerStrategy) GetName() string {
	return "Bollinger Bands (R)"
}
