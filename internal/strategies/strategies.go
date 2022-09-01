package strategies

import "tinkoff-invest-contest/internal/utils"

type Strategy interface {
	GetTradeSignal(instrument utils.InstrumentInterface, marketData MarketData, ordersConfig OrdersConfig) (*TradeSignal, map[string]any)
	GetOutputKeys() []string
	GetYAML() string
	GetName() string
}

var Names = make([]string, 0)

var JSONConstructors = make(map[string]func(string) (Strategy, error))
var DefaultsJSON = make(map[string]func() string)
