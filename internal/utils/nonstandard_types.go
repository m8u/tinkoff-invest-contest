package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"tinkoff-invest-contest/internal/client/investapi"
)

func FloatToMoneyValue(currency string, value float64) *investapi.MoneyValue {
	q := FloatToQuotation(value)
	return &investapi.MoneyValue{
		Currency: currency,
		Units:    q.Units,
		Nano:     q.Nano,
	}
}

func MoneyValueToFloat(m *investapi.MoneyValue) float64 {
	return QuotationToFloat(&investapi.Quotation{
		Units: m.Units,
		Nano:  m.Nano,
	})
}

func FloatToQuotation(value float64) *investapi.Quotation {
	split := strings.Split(fmt.Sprintf("%.9f", value), ".")
	units, err := strconv.ParseInt(split[0], 10, 64)
	MaybeCrash(err)
	nano, err := strconv.ParseInt(split[1], 10, 32)
	MaybeCrash(err)
	if value < 0 {
		nano *= -1
	}
	return &investapi.Quotation{
		Units: units,
		Nano:  int32(nano),
	}
}

func QuotationToFloat(q *investapi.Quotation) float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%+d", q.Units)+
		"."+
		fmt.Sprintf("%09d", int64(math.Abs(float64(q.Nano)))), 64)
	MaybeCrash(err)
	if q.Units == 0 && q.Nano < 0 {
		f *= -1
	}
	return f
}

func RoundQuotation(q, minPriceIncrement *investapi.Quotation) *investapi.Quotation {
	digits := len(strings.Split(
		strings.TrimRight(
			fmt.Sprintf(
				"%.9f", QuotationToFloat(minPriceIncrement),
			), "0"), ".")[1])
	f, err := strconv.ParseFloat(fmt.Sprintf("%.*f", digits, QuotationToFloat(q)), 64)
	MaybeCrash(err)
	return FloatToQuotation(f)
}
