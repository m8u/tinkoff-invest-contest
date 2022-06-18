package utils

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
)

// MaybeCrash выводит подробности об ошибке и завершает программу с кодом 1
// если ошибка != nil
func MaybeCrash(err error) {
	if err != nil {
		_, filename, line, _ := runtime.Caller(1)
		log.Fatalf("[error] %s:%d %v", filename, line, err)
	}
}

// WaitForInternetConnection пингует clients3.google.com, блокируя тред до успешного соединения
func WaitForInternetConnection() {
	httpClient := http.Client{Timeout: 5 * time.Second}
	err := errors.New("")
	for err != nil {
		_, err = httpClient.Get("https://clients3.google.com/")
		if err != nil {
			if !appstate.NoInternetConnection {
				log.Println("waiting for internet connection...")
			}
			appstate.NoInternetConnection = true
			time.Sleep(10 * time.Second)
		}
	}
	if appstate.NoInternetConnection {
		log.Println("internet connection established")
	}
	appstate.NoInternetConnection = false
}

func GetSandboxToken() string {
	token := os.Getenv("SANDBOX_TOKEN")
	if token == "" {
		log.Fatalln("please provide sandbox token via 'SANDBOX_TOKEN' environment variable")
	}
	return token
}

func GetCombatToken() string {
	token := os.Getenv("COMBAT_TOKEN")
	if token == "" {
		log.Fatalln("please provide combat token via 'COMBAT_TOKEN' environment variable")
	}
	return token
}

func GetGrafanaToken() string {
	token := os.Getenv("GRAFANA_TOKEN")
	if token == "" {
		log.Fatalln("please provide Grafana admin token via 'GRAFANA_TOKEN' environment variable")
	}
	return token
}

var CandleIntervalsV1NamesToValues = map[string]investapi.CandleInterval{
	"1min":  investapi.CandleInterval_CANDLE_INTERVAL_1_MIN,
	"5min":  investapi.CandleInterval_CANDLE_INTERVAL_5_MIN,
	"15min": investapi.CandleInterval_CANDLE_INTERVAL_15_MIN,
	"hour":  investapi.CandleInterval_CANDLE_INTERVAL_HOUR,
	"day":   investapi.CandleInterval_CANDLE_INTERVAL_DAY,
}
var CandleIntervalsValuesToV1Names = map[investapi.CandleInterval]string{
	investapi.CandleInterval_CANDLE_INTERVAL_1_MIN:  "1min",
	investapi.CandleInterval_CANDLE_INTERVAL_5_MIN:  "5min",
	investapi.CandleInterval_CANDLE_INTERVAL_15_MIN: "15min",
	investapi.CandleInterval_CANDLE_INTERVAL_HOUR:   "hour",
	investapi.CandleInterval_CANDLE_INTERVAL_DAY:    "day",
}

var CandleIntervalsToDurations = map[investapi.CandleInterval]time.Duration{
	investapi.CandleInterval_CANDLE_INTERVAL_1_MIN:  time.Minute,
	investapi.CandleInterval_CANDLE_INTERVAL_5_MIN:  5 * time.Minute,
	investapi.CandleInterval_CANDLE_INTERVAL_15_MIN: 15 * time.Minute,
	investapi.CandleInterval_CANDLE_INTERVAL_HOUR:   time.Hour,
	investapi.CandleInterval_CANDLE_INTERVAL_DAY:    24 * time.Hour,
}

type TradeSignal struct {
	Direction investapi.OrderDirection
}

func MoneyValueFromFloat(currency string, value float64) *investapi.MoneyValue {
	units, nano := math.Modf(value)
	return &investapi.MoneyValue{
		Currency: currency,
		Units:    int64(units),
		Nano:     int32(nano),
	}
}

func FloatFromMoneyValue(m *investapi.MoneyValue) float64 {
	return float64(m.Units) + float64(m.Nano)/math.Pow(10, float64(len(fmt.Sprint(m.Nano))))
}

func QuotationFromFloat(value float64) *investapi.Quotation {
	units, nano := math.Modf(value)
	return &investapi.Quotation{
		Units: int64(units),
		Nano:  int32(nano),
	}
}

func FloatFromQuotation(q *investapi.Quotation) float64 {
	return float64(q.Units) + float64(q.Nano)/math.Pow(10, float64(len(fmt.Sprint(q.Nano))))
}

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
