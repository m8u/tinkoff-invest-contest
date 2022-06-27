package tradeenv

import "tinkoff-invest-contest/internal/client/investapi"

func (e *TradeEnv) DoOrder(figi string, quantity int64, price float64, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType) error {
	err := e.Client.WrapOrder(e.isSandbox, figi, quantity, price, direction, accountId, orderType)
	if err != nil {
		return err
	}
	return nil
}
