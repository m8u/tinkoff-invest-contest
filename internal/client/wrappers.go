package client

import (
	"github.com/google/uuid"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func (c *Client) InstrumentByFigi(figi string, instrumentType utils.InstrumentType) (utils.InstrumentInterface, error) {
	var instrument utils.InstrumentInterface
	var err error
	switch instrumentType {
	case utils.InstrumentType_INSTRUMENT_TYPE_BOND:
		instrument, err = c.BondBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	case utils.InstrumentType_INSTRUMENT_TYPE_CURRENCY:
		instrument, err = c.CurrencyBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	case utils.InstrumentType_INSTRUMENT_TYPE_ETF:
		instrument, err = c.EtfBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	case utils.InstrumentType_INSTRUMENT_TYPE_FUTURE:
		instrument, err = c.FutureBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	case utils.InstrumentType_INSTRUMENT_TYPE_SHARE:
		instrument, err = c.ShareBy(investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI, "", figi)
	}
	if err != nil {
		return nil, err
	}
	return instrument, nil
}

// WrapOrder posts either sandbox or real order with automatically generated orderId and waits for order to be filled
func (c *Client) WrapOrder(isSandbox bool, figi string, quantity int64, price float64,
	direction investapi.OrderDirection, accountId string, orderType investapi.OrderType) (*investapi.PostOrderResponse, error) {
	var order *investapi.PostOrderResponse
	var err error
	if isSandbox {
		order, err = c.PostSandboxOrder(figi, quantity, price, direction, accountId, orderType, uuid.New().String())
		if err != nil {
			return order, err
		}
		status := order.ExecutionReportStatus
		for status != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
			orderState, err := c.GetSandboxOrderState(accountId, order.OrderId)
			if err != nil {
				return order, err
			}
			status = orderState.ExecutionReportStatus
		}
	} else {
		order, err = c.PostOrder(figi, quantity, price, direction, accountId, orderType, uuid.New().String())
		if err != nil {
			return order, err
		}
		status := order.ExecutionReportStatus
		for status != investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL {
			orderState, err := c.GetOrderState(accountId, order.OrderId)
			if err != nil {
				return order, err
			}
			status = orderState.ExecutionReportStatus
		}
	}
	return order, err
}
