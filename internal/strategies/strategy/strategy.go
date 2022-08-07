package strategy

import (
	"tinkoff-invest-contest/internal/strategies/strategy/bollinger"
	"tinkoff-invest-contest/internal/utils"
)

type Strategy interface {
	GetTradeSignal(marketData MarketData) (*utils.TradeSignal, map[string]any)
	GetOutputKeys() []string
	GetYAML() string
	GetName() string
}

var Names = [...]string{
	"bollinger",
}

var JSONConstructors map[string]func(string) (Strategy, error)
var DefaultsJSON map[string]func() string

func init() {
	JSONConstructors = map[string]func(string) (Strategy, error){
		"bollinger": bollinger.NewFromJSON,
	}
	DefaultsJSON = map[string]func() string{
		"bollinger": bollinger.GetDefaultsJSON,
	}
}
