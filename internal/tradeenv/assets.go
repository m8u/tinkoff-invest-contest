package tradeenv

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func (e *TradeEnv) CalculateMaxDealValue(accountId string, direction investapi.OrderDirection,
	instrument utils.InstrumentInterface, price *investapi.Quotation, allowMargin bool) (float64, error) {
	var positions *investapi.PositionsResponse
	var err error
	if e.isSandbox {
		positions, err = e.Client.GetSandboxPositions(accountId)
	} else {
		positions, err = e.Client.GetSandboxPositions(accountId)
	}
	if err != nil {
		return 0, err
	}

	var moneyHave float64
	var lotsHave int64
	for _, money := range positions.Money {
		if money.Currency == instrument.GetCurrency() {
			moneyHave = utils.FloatFromMoneyValue(money)
		}
	}
	if len(positions.Securities) > 0 {
		lotsHave = positions.Securities[0].Balance / int64(instrument.GetLot())
	} else {
		lotsHave = 0
	}

	var marginAttributes *investapi.GetMarginAttributesResponse
	if !e.isSandbox {
		marginAttributes, err = e.Client.GetMarginAttributes(accountId)
		if err != nil {
			return 0, err
		}
	}

	var maxDealValue float64
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		if allowMargin {
			liquidPortfolio := utils.FloatFromMoneyValue(marginAttributes.LiquidPortfolio)
			startMargin := utils.FloatFromMoneyValue(marginAttributes.StartingMargin)
			maxDealValue = (liquidPortfolio - startMargin) / utils.FloatFromQuotation(instrument.GetDlong())
		} else {
			maxDealValue = moneyHave
		}
		break
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		if allowMargin && instrument.GetShortEnabledFlag() {
			liquidPortfolio := utils.FloatFromMoneyValue(marginAttributes.LiquidPortfolio)
			startMargin := utils.FloatFromMoneyValue(marginAttributes.StartingMargin)
			maxDealValue = (liquidPortfolio - startMargin) / utils.FloatFromQuotation(instrument.GetDshort())
		} else {
			maxDealValue = float64(lotsHave) * float64(instrument.GetLot()) * utils.FloatFromQuotation(price)
		}
	}
	return maxDealValue, nil
}

func (e *TradeEnv) CalculateLotsCanAfford(direction investapi.OrderDirection, maxDealValue float64,
	instrument utils.InstrumentInterface, price *investapi.Quotation) int64 {

	priceFeeIncluded := utils.FloatFromQuotation(price)
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		priceFeeIncluded *= 1 + e.fee
		break
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		priceFeeIncluded *= 1 - e.fee
		break
	}

	lots := int64(maxDealValue / (priceFeeIncluded * float64(instrument.GetLot())))
	return lots
}

func (e *TradeEnv) GetLotsHave(accountId string, instrument utils.InstrumentInterface) (lots int64, err error) { // ???TODO: with expectation that it can return negative quantity for short position
	portfolio, err := e.Client.WrapGetPortfolio(e.isSandbox, accountId)
	if err != nil {
		return 0, err
	}
	for _, position := range portfolio.Positions {
		if position.Figi == instrument.GetFigi() {
			lots = int64(utils.FloatFromQuotation(position.QuantityLots))
		}
	}
	return
}
