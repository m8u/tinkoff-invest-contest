package tradeenv

import (
	"github.com/google/uuid"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

// DoOrder posts either sandbox or real order with automatically generated orderId and waits for order to be filled
func (e *TradeEnv) DoOrder(figi string, quantity int64, price *investapi.Quotation, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType) (avgPositionPrice float64, err error) {
	order, err := e.Client.WrapPostOrder(e.isSandbox, figi, quantity, price, direction, accountId, orderType, uuid.New().String())
	if err != nil {
		return
	}
	if e.isSandbox {
		orderState := new(investapi.OrderState)
		for orderState.ExecutionReportStatus != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
			orderState, err = e.Client.WrapGetOrderState(e.isSandbox, accountId, order.OrderId)
			if err != nil {
				return
			}
			time.Sleep(time.Second)
		}
		avgPositionPrice = utils.MoneyValueToFloat(orderState.AveragePositionPrice)
	} else {
		var orderTrades *investapi.OrderTrades
	loop:
		for {
			trades := e.GetTradesChannel(accountId)
			select {
			case orderTrades = <-trades:
				if orderTrades.OrderId == order.OrderId {
					break loop
				}
			}
		}
		var quantitySum int64
		for _, trade := range orderTrades.Trades {
			avgPositionPrice += utils.QuotationToFloat(trade.Price)
			quantitySum += trade.Quantity
		}
		avgPositionPrice /= float64(quantitySum)
	}
	return
}
