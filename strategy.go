package main

import (
	"math"
	investapi "tinkoff-invest-contest/investAPI"
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
func bollinger(candles []*investapi.HistoricCandle, coef float64) (float64, float64) {
	var sum float64
	n := float64(len(candles))

	// вычисляем арифметическое среднее "типичных" цен
	sum = 0
	for _, candle := range candles {
		sum += (FloatFromQuotation(candle.High) +
			FloatFromQuotation(candle.Low) +
			FloatFromQuotation(candle.Close)) / 3
	}
	mean := sum / n

	// вычисляем стандартное отклонение
	sum = 0
	for _, candle := range candles {
		sum += math.Pow(FloatFromQuotation(candle.Close)-mean, 2)
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

// isBetweenIncl определеяет, находится ли точка samplePoint включительно
// между точками bound1 и bound2
func isBetweenIncl(samplePoint float64, bound1 float64, bound2 float64) bool {
	return bound1 <= samplePoint && samplePoint <= bound2 ||
		bound2 <= samplePoint && samplePoint <= bound1
}

// GetTradeSignal формирует рекомендацию в виде торгового сигнала (TradeSignal)
func GetTradeSignal(strategyParams StrategyParams, testMode bool, currentCandle *investapi.Candle, newCandle bool, charts *Charts) *TradeSignal {
	lowerBound, upperBound := bollinger(
		(*charts.Candles)[len(*charts.Candles)-strategyParams.Window:len(*charts.Candles)-1],
		strategyParams.BollingerCoef,
	)
	// Добавляем интервал в статистику для отображения на графике
	// В режиме теста, если впервые, добавляем его дважды чтобы не словить index error при проверке пересечения
	if testMode && len(*charts.Intervals) == 0 {
		*charts.Intervals = append(*charts.Intervals, []float64{lowerBound, upperBound})
	}
	if newCandle {
		*charts.Intervals = append(*charts.Intervals, []float64{lowerBound, upperBound})
	}

	// Сигнал к покупке
	if isAroundPoint(FloatFromQuotation(currentCandle.Close), lowerBound, strategyParams.IntervalPointDeviation) ||
		(testMode && // в тестовом режиме также проверяем топорным способом
			(isBetweenIncl(
				lowerBound,
				FloatFromQuotation((*charts.Candles)[len(*charts.Candles)-2].Close),
				FloatFromQuotation(currentCandle.Close),
			) ||
				isBetweenIncl(
					lowerBound,
					FloatFromQuotation(currentCandle.Open),
					FloatFromQuotation(currentCandle.Close),
				))) {

		return &TradeSignal{investapi.OrderDirection_ORDER_DIRECTION_BUY}
		// Сигнал к продаже
	} else if isAroundPoint(FloatFromQuotation(currentCandle.Close), upperBound, strategyParams.IntervalPointDeviation) ||
		(testMode &&
			(isBetweenIncl(
				upperBound,
				FloatFromQuotation((*charts.Candles)[len(*charts.Candles)-2].Close),
				FloatFromQuotation(currentCandle.Close),
			) ||
				isBetweenIncl(
					upperBound,
					FloatFromQuotation(currentCandle.Open),
					FloatFromQuotation(currentCandle.Close),
				))) {

		return &TradeSignal{investapi.OrderDirection_ORDER_DIRECTION_SELL}
	}

	return nil
}
