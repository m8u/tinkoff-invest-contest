package main

import (
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"math"
)

type StrategyParams struct {
	// Количество последних свечей, на основе которых рассчитывается Bollinger Bands
	Window int
	// Коэффициент Bollinger Bands (множитель стандартного отклонения).
	BollingerCoef float64
	// Допустимое отклонение при проверке нахождения цены
	// в окресности границы интервала в процентах, например 0.001 (0.1%)
	IntervalPointDeviation float64
}

// bollinger вычисляет границы интервала Bollinger Bands
func bollinger(candles []sdk.Candle, coef float64) (float64, float64) {
	var sum float64
	n := float64(len(candles))

	// вычисляем арифметическое среднее "типичных" цен
	sum = 0
	for _, candle := range candles {
		sum += (candle.HighPrice + candle.LowPrice + candle.ClosePrice) / 3
	}
	mean := sum / n

	// вычисляем стандартное отклонение
	sum = 0
	for _, candle := range candles {
		sum += math.Pow(candle.ClosePrice-mean, 2)
	}
	sd := math.Sqrt(sum / n)

	// вычисляем границы интервала
	lowerBound := mean - coef*(sd)
	upperBound := mean + coef*(sd)

	return lowerBound, upperBound
}

// isAroundPoint определяет, находится ли точка samplePoint в окрестности
// точки refPoint с допустимым отклонением deviation
func isAroundPoint(samplePoint float64, refPoint float64, deviation float64) bool {
	return samplePoint >= refPoint-refPoint*deviation &&
		samplePoint <= refPoint+refPoint*deviation
}

// GetTradeSignal формирует рекомендацию в виде торгового сигнала (TradeSignal)
func GetTradeSignal(strategyParams StrategyParams, testMode bool, currentCandle sdk.Candle, newCandle bool, charts *Charts) *TradeSignal {
	lowerBound, upperBound := bollinger(
		charts.Candles[len(charts.Candles)-strategyParams.Window:len(charts.Candles)-1],
		strategyParams.BollingerCoef,
	)
	// Добавляем интервал в статистику для отображения на графике
	// В режиме теста, если впервые, добавляем его дважды чтобы не словить index error при проверке пересечения
	if testMode && len(charts.Intervals) == 0 {
		charts.Intervals = append(charts.Intervals, []float64{lowerBound, upperBound})
	}
	if newCandle {
		charts.Intervals = append(charts.Intervals, []float64{lowerBound, upperBound})
	}

	if isAroundPoint(currentCandle.ClosePrice, lowerBound, strategyParams.IntervalPointDeviation) ||
		(testMode && // в тестовом режиме также проверяем топорным способом
			((charts.Candles[len(charts.Candles)-2].ClosePrice <= charts.Intervals[len(charts.Intervals)-2][0] &&
				currentCandle.ClosePrice >= lowerBound) ||
				(charts.Candles[len(charts.Candles)-2].ClosePrice >= charts.Intervals[len(charts.Intervals)-2][0] &&
					currentCandle.ClosePrice <= lowerBound))) { // buy сигнал

		return &TradeSignal{sdk.BUY}

	} else if isAroundPoint(currentCandle.ClosePrice, upperBound, strategyParams.IntervalPointDeviation) ||
		(testMode &&
			((charts.Candles[len(charts.Candles)-2].ClosePrice >= charts.Intervals[len(charts.Intervals)-2][1] &&
				currentCandle.ClosePrice <= upperBound) ||
				(charts.Candles[len(charts.Candles)-2].ClosePrice <= charts.Intervals[len(charts.Intervals)-2][1] &&
					currentCandle.ClosePrice >= upperBound))) { // sell сигнал

		return &TradeSignal{sdk.SELL}
	}

	return nil
}
