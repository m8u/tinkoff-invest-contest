package strategies

import (
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	"tinkoff-invest-contest/internal/strategies/tistrategy/bollinger"
)

func InitConstructorsMap() {
	tistrategy.JsonConstructors = map[string]func(string) (tistrategy.TechnicalIndicatorStrategy, error){
		"bollinger": bollinger.NewFromJsonString,
	}
}
