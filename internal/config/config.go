package config

import "tinkoff-invest-contest/internal/client/investapi"

type Config struct {
	IsSandbox   bool
	Token       string
	NumAccounts int
	Money       float64
	Fee         float64
	Instruments []ConfigInstrument
}

type ConfigInstrument struct {
	FIGI           string
	CandleInterval investapi.CandleInterval
	OrderBookDepth int32
}
