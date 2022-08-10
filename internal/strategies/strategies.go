package strategies

import (
	"tinkoff-invest-contest/internal/utils"
)

type Strategy interface {
	GetTradeSignal(marketData MarketData) (*utils.TradeSignal, map[string]any)
	GetOutputKeys() []string
	GetYAML() string
	GetName() string
}

var Names = make([]string, 0)

var JSONConstructors = make(map[string]func(string) (Strategy, error))
var DefaultsJSON = make(map[string]func() string)
