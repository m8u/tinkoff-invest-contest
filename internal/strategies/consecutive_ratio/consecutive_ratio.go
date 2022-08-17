/*
consecutive_ratio.go describes a strategy that generates
a trade signal when the ratio of asks or bids to all orders
in order book satisfies the condition (ratio >= triggerRatio)
for specified amount of times.
*/

package consecutive_ratio

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
	"log"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies"
	"tinkoff-invest-contest/internal/utils"
)

type consecutiveRatioStrategy struct {
	triggerRatio         float64 // trigger ratio of bids or asks against all orders
	triggerTimesRepeated int     // times repeated that counts as a trade signal

	flag          investapi.OrderDirection // currently awaited trade signal (based on previous ratios)
	timesRepeated int                      // current times repeated
}

type consecutiveRatioParams struct {
	Ratio         float64 `json:"ratio" yaml:"Ratio"`
	TimesRepeated int     `json:"timesRepeated" yaml:"TimesRepeated"`
}

func init() {
	strategyName := "consecutive_ratio"
	strategies.Names = append(strategies.Names, strategyName)
	strategies.JSONConstructors[strategyName] = NewFromJSON
	strategies.DefaultsJSON[strategyName] = GetDefaultsJSON
}

func NewFromJSON(s string) (strategies.Strategy, error) {
	p := consecutiveRatioParams{}

	err := json.Unmarshal([]byte(s), &p)
	if err != nil {
		return nil, err
	}

	return &consecutiveRatioStrategy{
		triggerRatio:         p.Ratio,
		triggerTimesRepeated: p.TimesRepeated,
	}, nil
}

func GetDefaultsJSON() string {
	defaults := consecutiveRatioParams{
		Ratio:         0.8,
		TimesRepeated: 10,
	}
	bytes, err := json.MarshalIndent(&defaults, "", "  ")
	utils.MaybeCrash(err)
	return string(bytes)
}

func (s *consecutiveRatioStrategy) GetTradeSignal(marketData strategies.MarketData) (*utils.TradeSignal, map[string]any) {
	var bids, asks int64
	for _, order := range marketData.OrderBook.Bids {
		bids += order.Quantity
	}
	for _, order := range marketData.OrderBook.Asks {
		asks += order.Quantity
	}

	bidsRatio, asksRatio := float64(bids)/float64(asks+bids), float64(asks)/float64(asks+bids)

	if bidsRatio >= s.triggerRatio && s.flag != investapi.OrderDirection_ORDER_DIRECTION_SELL {
		s.flag = investapi.OrderDirection_ORDER_DIRECTION_BUY
		s.timesRepeated++
	} else if asksRatio >= s.triggerRatio && s.flag != investapi.OrderDirection_ORDER_DIRECTION_BUY {
		s.flag = investapi.OrderDirection_ORDER_DIRECTION_SELL
		s.timesRepeated++
	} else {
		s.flag = investapi.OrderDirection_ORDER_DIRECTION_UNSPECIFIED
		s.timesRepeated = 0
	}

	log.Println(marketData.OrderBook.Figi, bidsRatio, asksRatio, s.timesRepeated)

	var signal *utils.TradeSignal
	if s.timesRepeated >= s.triggerTimesRepeated {
		signal = &utils.TradeSignal{
			Direction: s.flag,
		}
	}

	outputValues := map[string]any{}

	return signal, outputValues
}

func (*consecutiveRatioStrategy) GetOutputKeys() []string {
	return []string{}
}

func (s *consecutiveRatioStrategy) GetYAML() string {
	obj := consecutiveRatioParams{
		Ratio:         s.triggerRatio,
		TimesRepeated: s.triggerTimesRepeated,
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}

func (*consecutiveRatioStrategy) GetName() string {
	return "Consecutive ratio"
}
