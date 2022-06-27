package utils

import "tinkoff-invest-contest/internal/client/investapi"

type InstrumentInterface interface {
	GetDlong() *investapi.Quotation
	GetDshort() *investapi.Quotation
	GetLot() int32
	GetShortEnabledFlag() bool
}

type InstrumentType int32

const (
	InstrumentType_INSTRUMENT_TYPE_BOND     InstrumentType = 0
	InstrumentType_INSTRUMENT_TYPE_CURRENCY InstrumentType = 1
	InstrumentType_INSTRUMENT_TYPE_ETF      InstrumentType = 2
	InstrumentType_INSTRUMENT_TYPE_FUTURE   InstrumentType = 3
	InstrumentType_INSTRUMENT_TYPE_SHARE    InstrumentType = 4
)
