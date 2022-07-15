package strategies

import (
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	"tinkoff-invest-contest/internal/strategies/tistrategy/bollinger"
)

var Names = [...]string{
	"bollinger",
}

func init() {
	tistrategy.JSONConstructors = map[string]func(string) (tistrategy.TechnicalIndicatorStrategy, error){
		"bollinger": bollinger.NewFromJSON,
	}
	tistrategy.DefaultsJSON = map[string]func() string{
		"bollinger": bollinger.GetDefaultsJSON,
	}
}
