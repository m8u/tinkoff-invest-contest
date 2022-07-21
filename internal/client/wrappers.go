package client

import (
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

func (c *Client) WrapPostOrder(isSandbox bool, figi string, quantity int64, price *investapi.Quotation,
	direction investapi.OrderDirection, accountId string, orderType investapi.OrderType, orderId string) (*investapi.PostOrderResponse, error) {
	var order *investapi.PostOrderResponse
	var err error
	if isSandbox {
		order, err = c.PostSandboxOrder(figi, quantity, utils.QuotationToFloat(price), direction, accountId, orderType, orderId)
		if err != nil {
			return nil, err
		}
	} else {
		order, err = c.PostOrder(figi, quantity, utils.QuotationToFloat(price), direction, accountId, orderType, orderId)
		if err != nil {
			return nil, err
		}
	}
	return order, nil
}

func (c *Client) WrapGetOrderState(isSandbox bool, accountId string, orderId string) (*investapi.OrderState, error) {
	var state *investapi.OrderState
	var err error
	if isSandbox {
		state, err = c.GetSandboxOrderState(accountId, orderId)
		if err != nil {
			return nil, err
		}
	} else {
		state, err = c.GetOrderState(accountId, orderId)
		if err != nil {
			return nil, err
		}
	}
	return state, nil
}

func (c *Client) WrapGetPortfolio(isSandbox bool, accountId string) (*investapi.PortfolioResponse, error) {
	var portfolio *investapi.PortfolioResponse
	var err error
	if isSandbox {
		portfolio, err = c.GetSandboxPortfolio(accountId)
	} else {
		portfolio, err = c.GetPortfolio(accountId)
	}
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (c *Client) WrapGetPositions(isSandbox bool, accountId string) (*investapi.PositionsResponse, error) {
	var positions *investapi.PositionsResponse
	var err error
	if isSandbox {
		positions, err = c.GetSandboxPositions(accountId)
	} else {
		positions, err = c.GetPositions(accountId)
	}
	if err != nil {
		return nil, err
	}
	return positions, nil
}
