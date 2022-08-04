package indicators

import (
	"math"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

type BollingerBands struct {
	// Bollinger Bands coefficient (multiple of standard deviation)
	coef float64
}

func NewBollingerBands(coef float64) *BollingerBands {
	return &BollingerBands{coef: coef}
}

// Calculate calculates Bollinger Bands interval bounds
func (bollinger *BollingerBands) Calculate(candles []*investapi.HistoricCandle) (float64, float64) {
	var sum float64
	n := float64(len(candles))

	// calculate an average of "typical" prices
	sum = 0
	for _, candle := range candles {
		sum += (utils.QuotationToFloat(candle.High) +
			utils.QuotationToFloat(candle.Low) +
			utils.QuotationToFloat(candle.Close)) / 3
	}
	mean := sum / n

	// calculate standard deviation
	sum = 0
	for _, candle := range candles {
		sum += math.Pow(utils.QuotationToFloat(candle.Close)-mean, 2)
	}
	sd := math.Sqrt(sum / n)

	// calculate interval bounds
	lowerBound := mean - bollinger.coef*(sd)
	upperBound := mean + bollinger.coef*(sd)

	return lowerBound, upperBound
}

func (bollinger *BollingerBands) GetCoef() float64 {
	return bollinger.coef
}
