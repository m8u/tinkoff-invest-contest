package utils

import (
	"fmt"
	"tinkoff-invest-contest/internal/client/investapi"
)

func StringToCandleInterval(s string) (investapi.CandleInterval, error) {
	switch s {
	case "1min":
		return investapi.CandleInterval_CANDLE_INTERVAL_1_MIN, nil
	case "5min":
		return investapi.CandleInterval_CANDLE_INTERVAL_5_MIN, nil
	case "15min":
		return investapi.CandleInterval_CANDLE_INTERVAL_15_MIN, nil
	case "1hour":
		return investapi.CandleInterval_CANDLE_INTERVAL_HOUR, nil
	case "1day":
		return investapi.CandleInterval_CANDLE_INTERVAL_DAY, nil
	default:
		return investapi.CandleInterval_CANDLE_INTERVAL_UNSPECIFIED,
			fmt.Errorf("unknown candle interval: %q", s)
	}
}

func CandleIntervalToString(interval investapi.CandleInterval) string {
	switch interval {
	case investapi.CandleInterval_CANDLE_INTERVAL_1_MIN:
		return "1min"
	case investapi.CandleInterval_CANDLE_INTERVAL_5_MIN:
		return "5min"
	case investapi.CandleInterval_CANDLE_INTERVAL_15_MIN:
		return "15min"
	case investapi.CandleInterval_CANDLE_INTERVAL_HOUR:
		return "1hour"
	case investapi.CandleInterval_CANDLE_INTERVAL_DAY:
		return "1day"
	}
	return ""
}
