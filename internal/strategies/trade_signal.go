package strategies

import (
	"math"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type TradeSignalOrder struct {
	Direction investapi.OrderDirection
	Type      investapi.OrderType
	Price     *investapi.Quotation
}

type TradeSignal struct {
	Order                *TradeSignalOrder
	StopLoss, TakeProfit *TradeSignalStopOrder
}

func NewTradeSignal(direction investapi.OrderDirection, orderType investapi.OrderType,
	price *investapi.Quotation) *TradeSignal {
	return &TradeSignal{
		Order: &TradeSignalOrder{
			Direction: direction,
			Type:      orderType,
			Price:     price,
		},
	}
}

func NewTradeSignalWithStopOrders(direction investapi.OrderDirection, orderType investapi.OrderType,
	price *investapi.Quotation, stopLossOrderType investapi.OrderType,
	profitTriggerRatio, lossTriggerRatio, lossLimitRatio float64, minPriceIncrement *investapi.Quotation) *TradeSignal {
	stopOrdersDirection := utils.ReverseOrderDirection(direction)
	signal := &TradeSignal{
		Order: &TradeSignalOrder{
			Direction: direction,
			Type:      orderType,
			Price:     price,
		},
		TakeProfit: &TradeSignalStopOrder{
			Direction: stopOrdersDirection,
			Type:      investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT,
		},
		StopLoss: &TradeSignalStopOrder{
			Direction: stopOrdersDirection,
		},
	}
	priceFloat := utils.QuotationToFloat(price)
	signal.TakeProfit.TriggerPrice = utils.RoundQuotation(utils.FloatToQuotation(
		priceFloat*(1+profitTriggerRatio*math.Pow(-1, float64(stopOrdersDirection))),
	), minPriceIncrement)

	signal.StopLoss.TriggerPrice = utils.RoundQuotation(utils.FloatToQuotation(
		priceFloat*(1-lossTriggerRatio*math.Pow(-1, float64(stopOrdersDirection))),
	), minPriceIncrement)
	if stopLossOrderType == investapi.OrderType_ORDER_TYPE_LIMIT {
		signal.StopLoss.LimitPrice = utils.RoundQuotation(utils.FloatToQuotation(
			priceFloat*(1-lossLimitRatio*math.Pow(-1, float64(stopOrdersDirection))),
		), minPriceIncrement)
		signal.StopLoss.Type = investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LIMIT
	}

	return signal
}
