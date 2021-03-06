package tradeenv

import (
	"github.com/google/uuid"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
)

// DoOrder posts either sandbox or real order with automatically generated orderId and waits for order to be filled
func (e *TradeEnv) DoOrder(figi string, quantity int64, price *investapi.Quotation, direction investapi.OrderDirection,
	accountId string, orderType investapi.OrderType) error {
	order, err := e.Client.WrapPostOrder(e.isSandbox, figi, quantity, price, direction, accountId, orderType, uuid.New().String())
	if err != nil {
		return err
	}
	status := order.ExecutionReportStatus
	for status != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
		orderState, err := e.Client.WrapGetOrderState(e.isSandbox, accountId, order.OrderId)
		if err != nil {
			return err
		}
		status = orderState.ExecutionReportStatus
		time.Sleep(time.Second)
	}
	return err
}
