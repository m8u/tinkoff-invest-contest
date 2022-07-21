package utils

type Tariff string

const (
	Investor Tariff = "investor"
	Trader   Tariff = "trader"
	Premium  Tariff = "premium"
)

var Fees = map[Tariff]float64{
	Investor: 0.003,
	Trader:   0.0004,
	Premium:  0.00025,
}
