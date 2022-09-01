package strategies

import "tinkoff-invest-contest/internal/client/investapi"

type OrdersConfig struct {
	OrderType         investapi.OrderType
	StopLossOrderType investapi.OrderType
	TakeProfitRatio   float64
	StopLossRatio     float64
	StopLossExecRatio float64
}
