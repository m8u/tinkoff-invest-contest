package utils

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
)

// MaybeCrash выводит подробности об ошибке и завершает программу с кодом 1
// если ошибка != nil
func MaybeCrash(err error) {
	if err != nil {
		log.Fatalln(PrettifyError(err))
	}
}

func PrettifyError(err error) string {
	_, filename, line, _ := runtime.Caller(1)
	return fmt.Sprintf("[error] %s:%d %v", filename, line, err)
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
		log.Println("please provide Grafana admin token via 'GRAFANA_TOKEN' environment variable")
	}
	return token
}

type TradeSignal struct {
	Direction investapi.OrderDirection
}

func OrderDirectionToString(direction investapi.OrderDirection) string {
	switch direction {
	case investapi.OrderDirection_ORDER_DIRECTION_BUY:
		return "BUY"
	case investapi.OrderDirection_ORDER_DIRECTION_SELL:
		return "SELL"
	default:
		return ""
	}
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
	f, err := strconv.ParseFloat(fmt.Sprintf("%+d", m.Units)+
		"."+
		fmt.Sprintf("%09d", int64(math.Abs(float64(m.Nano)))), 64)
	MaybeCrash(err)
	return f
}

func QuotationFromFloat(value float64) *investapi.Quotation {
	units, nano := math.Modf(value)
	return &investapi.Quotation{
		Units: int64(units),
		Nano:  int32(nano),
	}
}

func FloatFromQuotation(q *investapi.Quotation) float64 {
	f, err := strconv.ParseFloat(fmt.Sprintf("%+d", q.Units)+
		"."+
		fmt.Sprintf("%09d", int64(math.Abs(float64(q.Nano)))), 64)
	MaybeCrash(err)
	return f
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
