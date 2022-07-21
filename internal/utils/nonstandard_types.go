package utils

import (
	"fmt"
	"math"
	"strconv"
	"tinkoff-invest-contest/internal/client/investapi"
)

func FloatToMoneyValue(currency string, value float64) *investapi.MoneyValue {
	units, nano := math.Modf(value)
	return &investapi.MoneyValue{
		Currency: currency,
		Units:    int64(units),
		Nano:     int32(nano),
	}
}

func MoneyValueToFloat(m *investapi.MoneyValue) float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%+d", m.Units)+
		"."+
		fmt.Sprintf("%09d", int64(math.Abs(float64(m.Nano)))), 64)
	MaybeCrash(err)
	return f
}

func FloatToQuotation(value float64) *investapi.Quotation {
	units, nano := math.Modf(value)
	return &investapi.Quotation{
		Units: int64(units),
		Nano:  int32(nano),
	}
}

func QuotationToFloat(q *investapi.Quotation) float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%+d", q.Units)+
		"."+
		fmt.Sprintf("%09d", int64(math.Abs(float64(q.Nano)))), 64)
	MaybeCrash(err)
	return f
}
