package utils

import (
	"fmt"
	"tinkoff-invest-contest/internal/client/investapi"
)

type InstrumentInterface interface {
	GetDlong() *investapi.Quotation
	GetDshort() *investapi.Quotation
	GetLot() int32
	GetShortEnabledFlag() bool
	GetCurrency() string
	GetFigi() string
}

type InstrumentType int32

const (
	InstrumentType_INSTRUMENT_TYPE_BOND     InstrumentType = 0
	InstrumentType_INSTRUMENT_TYPE_CURRENCY InstrumentType = 1
	InstrumentType_INSTRUMENT_TYPE_ETF      InstrumentType = 2
	InstrumentType_INSTRUMENT_TYPE_FUTURE   InstrumentType = 3
	InstrumentType_INSTRUMENT_TYPE_SHARE    InstrumentType = 4
)

func InstrumentTypeFromString(s string) (InstrumentType, error) {
	switch s {
	case "bond":
		return InstrumentType_INSTRUMENT_TYPE_BOND, nil
	case "currency":
		return InstrumentType_INSTRUMENT_TYPE_CURRENCY, nil
	case "etf":
		return InstrumentType_INSTRUMENT_TYPE_ETF, nil
	case "future":
		return InstrumentType_INSTRUMENT_TYPE_FUTURE, nil
	case "share":
		return InstrumentType_INSTRUMENT_TYPE_SHARE, nil
	default:
		return -1, fmt.Errorf("unknown instrument type: %q", s)
	}
}
