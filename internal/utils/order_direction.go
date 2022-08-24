package utils

import (
	"math"
	"tinkoff-invest-contest/internal/client/investapi"
)

func OrderDirectionToString(direction investapi.OrderDirection) string {
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		return "BUY"
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		return "SELL"
	default:
		return ""
	}
}

func ReverseOrderDirection(direction investapi.OrderDirection) investapi.OrderDirection {
	return direction + investapi.OrderDirection(math.Pow(-1, float64(direction+1)))
}
