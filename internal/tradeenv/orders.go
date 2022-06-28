package tradeenv

import "tinkoff-invest-contest/internal/client/investapi"

func (e *TradeEnv) DoOrder(figi string, quantity int64, price float64, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType) (*investapi.PostOrderResponse, error) {
	return e.Client.WrapOrder(e.isSandbox, figi, quantity, price, direction, accountId, orderType)
}
