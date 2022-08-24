package strategies

import (
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TradeSignalStopOrder struct {
	Direction    investapi.OrderDirection
	Type         investapi.StopOrderType
	TriggerPrice *investapi.Quotation
	LimitPrice   *investapi.Quotation
}

func (stopOrder *TradeSignalStopOrder) IsTriggered(price *investapi.Quotation) bool {
	priceFloat := utils.QuotationToFloat(price)
	triggerPriceFloat := utils.QuotationToFloat(stopOrder.TriggerPrice)
	if stopOrder.Type == investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT {
		if (stopOrder.Direction == investapi.OrderDirection_ORDER_DIRECTION_SELL &&
			priceFloat >= triggerPriceFloat) ||
			(stopOrder.Direction == investapi.OrderDirection_ORDER_DIRECTION_BUY &&
				priceFloat <= triggerPriceFloat) {
			return true
		}
	} else if (stopOrder.Direction == investapi.OrderDirection_ORDER_DIRECTION_SELL &&
		priceFloat <= triggerPriceFloat) ||
		(stopOrder.Direction == investapi.OrderDirection_ORDER_DIRECTION_BUY &&
			priceFloat >= triggerPriceFloat) {
		return true
	}
	return false
}
