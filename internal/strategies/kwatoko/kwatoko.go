/*
kwatoko.go describes an orderbook strategy from
https://github.com/AndreVasilev/Kwatoko. Its goal is to
earn from the price jumps that occur when an order with
anomalously high lots quantity appears in the orderbook.
*/

package kwatoko

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies"
	"tinkoff-invest-contest/internal/utils"
)

type kwatokoStrategy struct {
	anomalyThreshold float64 // trigger ratio of (anomalous order quantity):(avg. orders quantity)
	priceDelta       float64 // abs(new order price delta relative to anomalous order price)
}

type kwatokoParams struct {
	AnomalyThreshold float64 `json:"anomalyThreshold" yaml:"AnomalyThreshold"`
	PriceDelta       float64 `json:"priceDelta" yaml:"PriceDelta"`
}

func init() {
	strategyName := "kwatoko"
	strategies.Names = append(strategies.Names, strategyName)
	strategies.JSONConstructors[strategyName] = NewFromJSON
	strategies.DefaultsJSON[strategyName] = GetDefaultsJSON
}

func NewFromJSON(s string) (strategies.Strategy, error) {
	p := kwatokoParams{}

	err := json.Unmarshal([]byte(s), &p)
	if err != nil {
		return nil, err
	}

	return &kwatokoStrategy{
		anomalyThreshold: p.AnomalyThreshold,
		priceDelta:       p.PriceDelta,
	}, nil
}

func GetDefaultsJSON() string {
	defaults := kwatokoParams{
		AnomalyThreshold: 10,
		PriceDelta:       0.001,
	}
	bytes, err := json.MarshalIndent(&defaults, "", "  ")
	utils.MaybeCrash(err)
	return string(bytes)
}

func (s *kwatokoStrategy) GetTradeSignal(instrument utils.InstrumentInterface, marketData strategies.MarketData,
	ordersConfig strategies.OrdersConfig) (*strategies.TradeSignal, map[string]any) {

	var avgQuantity float64
	for _, bid := range marketData.OrderBook.Bids {
		avgQuantity += float64(bid.Quantity)
	}
	for _, ask := range marketData.OrderBook.Asks {
		avgQuantity += float64(ask.Quantity)
	}
	avgQuantity /= float64(len(marketData.OrderBook.Bids) + len(marketData.OrderBook.Asks))

	var signal *strategies.TradeSignal
	if float64(marketData.OrderBook.Bids[0].Quantity)/avgQuantity >= s.anomalyThreshold {
		deltedPrice := utils.RoundQuotation(
			utils.FloatToQuotation(utils.QuotationToFloat(marketData.OrderBook.Bids[0].Price)*(1+s.priceDelta)),
			instrument.GetMinPriceIncrement(),
		)
		signal = strategies.NewTradeSignalWithStopOrders(
			investapi.OrderDirection_ORDER_DIRECTION_BUY,
			deltedPrice,
			instrument.GetMinPriceIncrement(),
			ordersConfig,
		)
	} else if float64(marketData.OrderBook.Asks[0].Quantity)/avgQuantity >= s.anomalyThreshold {
		deltedPrice := utils.RoundQuotation(
			utils.FloatToQuotation(utils.QuotationToFloat(marketData.OrderBook.Asks[0].Price)*(1-s.priceDelta)),
			instrument.GetMinPriceIncrement(),
		)
		signal = strategies.NewTradeSignalWithStopOrders(
			investapi.OrderDirection_ORDER_DIRECTION_SELL,
			deltedPrice,
			instrument.GetMinPriceIncrement(),
			ordersConfig,
		)
	}

	return signal, map[string]any{}
}

func (*kwatokoStrategy) GetOutputKeys() []string {
	return []string{}
}

func (s *kwatokoStrategy) GetYAML() string {
	obj := kwatokoParams{
		AnomalyThreshold: s.anomalyThreshold,
		PriceDelta:       s.priceDelta,
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}

func (*kwatokoStrategy) GetName() string {
	return "Kwatoko"
}
