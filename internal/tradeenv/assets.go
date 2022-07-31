package tradeenv

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

func (e *TradeEnv) CalculateMaxDealValue(accountId string, direction investapi.OrderDirection,
	instrument utils.InstrumentInterface, price *investapi.Quotation, allowMargin bool) float64 {
	var positions *investapi.PositionsResponse
	var err error
	if e.isSandbox {
		positions, err = e.Client.GetSandboxPositions(accountId)
	} else {
		positions, err = e.Client.GetSandboxPositions(accountId)
	}
	utils.MaybeCrash(err)

	var moneyHave float64
	var lotsHave int64
	for _, money := range positions.Money {
		if money.Currency == instrument.GetCurrency() {
			moneyHave = utils.MoneyValueToFloat(money)
		}
	}
	for _, position := range positions.Securities {
		if position.Figi == instrument.GetFigi() {
			lotsHave = position.Balance / int64(instrument.GetLot())
		}
	}

	var marginAttributes *investapi.GetMarginAttributesResponse
	if !e.isSandbox {
		marginAttributes, err = e.Client.GetMarginAttributes(accountId)
		utils.MaybeCrash(err)
	}

	var maxDealValue float64
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		if allowMargin {
			liquidPortfolio := utils.MoneyValueToFloat(marginAttributes.LiquidPortfolio)
			startMargin := utils.MoneyValueToFloat(marginAttributes.StartingMargin)
			maxDealValue = (liquidPortfolio - startMargin) / utils.QuotationToFloat(instrument.GetDlong())
		} else {
			maxDealValue = moneyHave
		}
		break
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		if allowMargin && instrument.GetShortEnabledFlag() {
			liquidPortfolio := utils.MoneyValueToFloat(marginAttributes.LiquidPortfolio)
			startMargin := utils.MoneyValueToFloat(marginAttributes.StartingMargin)
			maxDealValue = (liquidPortfolio - startMargin) / utils.QuotationToFloat(instrument.GetDshort())
		} else {
			maxDealValue = float64(lotsHave) * float64(instrument.GetLot()) * utils.QuotationToFloat(price)
		}
	}
	return maxDealValue
}

func (e *TradeEnv) CalculateLotsCanAfford(direction investapi.OrderDirection, maxDealValue float64,
	instrument utils.InstrumentInterface, price *investapi.Quotation, fee float64) int64 {

	priceFeeIncluded := utils.QuotationToFloat(price)
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		priceFeeIncluded *= 1 + fee
		break
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		priceFeeIncluded *= 1 - fee
		break
	}

	lots := int64(maxDealValue / (priceFeeIncluded * float64(instrument.GetLot())))
	return lots
}

func (e *TradeEnv) GetLotsHave(accountId string, instrument utils.InstrumentInterface) (lots int64, err error) { // TODO: with expectation that it can return negative quantity for short position
	portfolio, err := e.Client.WrapGetPortfolio(e.isSandbox, accountId)
	if err != nil {
		return 0, err
	}
	for _, position := range portfolio.Positions {
		if position.Figi == instrument.GetFigi() {
			lots = int64(utils.QuotationToFloat(position.QuantityLots))
		}
	}
	return
}
