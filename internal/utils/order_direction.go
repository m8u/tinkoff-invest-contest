package utils

import (
	"tinkoff-invest-contest/internal/client/investapi"
)

type TradeSignal struct {
	Direction investapi.OrderDirection
}

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
