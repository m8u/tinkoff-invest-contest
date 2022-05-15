package main

import (
	"fmt"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"math"
	"runtime"
	"strconv"
	"time"
)

// MaybeCrash выводит подробности об ошибке и завершает программу с кодом 1
// если ошибка - не nil
func MaybeCrash(err error) {
	if err != nil {
		_, filename, line, _ := runtime.Caller(1)
		log.Fatalf("[error] %s:%d %v", filename, line, err)
	}
}

// GetCandlesForLastNDays загружает свечи за заданное кол-во последних дней
func GetCandlesForLastNDays(client *RestClientV2, figi string, n int, candleInterval sdk.CandleInterval) ([]sdk.Candle, error) {
	candles := make([]sdk.Candle, 0)
	for day := n; day >= 0; day-- {
		portion, err := client.GetCandles(
			time.Now().AddDate(0, 0, -day-1),
			time.Now().AddDate(0, 0, -day),
			candleInterval,
			figi,
		)
		if err != nil {
			return nil, err
		}
		for _, currentCandle := range portion {
			candles = append(candles, currentCandle)
		}
	}
	return candles, nil
}

var CandleIntervalsToDurations = map[sdk.CandleInterval]time.Duration{
	sdk.CandleInterval1Min: time.Minute,
	//sdk.CandleInterval2Min:  2 * time.Minute,
	//sdk.CandleInterval3Min:  3 * time.Minute,
	sdk.CandleInterval5Min: 5 * time.Minute,
	//sdk.CandleInterval10Min: 10 * time.Minute,
	sdk.CandleInterval15Min: 15 * time.Minute,
	//sdk.CandleInterval30Min: 30 * time.Minute,
	sdk.CandleInterval1Hour: time.Hour,
	//sdk.CandleInterval2Hour: 2 * time.Hour,
	//sdk.CandleInterval4Hour: 4 * time.Hour,
	sdk.CandleInterval1Day: 24 * time.Hour,
}

type TradeSignal struct {
	// sdk.BUY или sdk.SELL
	Direction sdk.OperationType
}

type MoneyValue struct {
	Currency string `json:"currency"`
	Units    string `json:"units"`
	Nano     int32  `json:"nano"`
}

func (m *MoneyValue) ToFloat() float64 {
	unitsNum, _ := strconv.Atoi(m.Units)
	return float64(unitsNum) + float64(m.Nano)/math.Pow(10, float64(len(fmt.Sprint(m.Nano))))
}

type Quotation struct {
	Units string
	Nano  int32
}

func (q *Quotation) ToFloat() float64 {
	unitsNum, _ := strconv.Atoi(q.Units)
	return float64(unitsNum) + float64(q.Nano)/math.Pow(10, float64(len(fmt.Sprint(q.Nano))))
}

// MarginAttributes - тело ответа метода GetMarginAttributes API версии 1
type MarginAttributes struct {
	LiquidPortfolio       MoneyValue `json:"liquidPortfolio"`
	StartingMargin        MoneyValue `json:"startingMargin"`
	MinimalMargin         MoneyValue `json:"minimalMargin"`
	FundsSufficiencyLevel MoneyValue `json:"fundsSufficiencyLevel"`
	AmountOfMissingFunds  MoneyValue `json:"amountOfMissingFunds"`
}

type Tariff string

const (
	Investor Tariff = "investor"
	Trader   Tariff = "trader"
	Premium  Tariff = "premium"
)

// UserInfo - тело ответа метода GetInfo API версии 1
type UserInfo struct {
	PremStatus           bool     `json:"premStatus"`
	QualStatus           bool     `json:"qualStatus"`
	QualifiedForWorkWith []string `json:"qualifiedForWorkWith"`
	Tariff               Tariff   `json:"tariff"`
}

var Fees = map[Tariff]float64{
	Investor: 0.003,
	Trader:   0.0004,
	Premium:  0.00025,
}

type InstrumentIdType int

const (
	UNSPECIFIED InstrumentIdType = 0
	FIGI        InstrumentIdType = 1
	TICKER      InstrumentIdType = 2
	UID         InstrumentIdType = 3
)

// Share - тело ответа метода ShareBy API версии 1
type Share struct {
	Instrument struct {
		Figi                  string     `json:"figi"`
		DshortMin             Quotation  `json:"dshortMin"`
		CountryOfRisk         string     `json:"countryOfRisk"`
		Lot                   int        `json:"lot"`
		Uid                   string     `json:"uid"`
		Dlong                 Quotation  `json:"dlong"`
		Nominal               MoneyValue `json:"nominal"`
		SellAvailableFlag     bool       `json:"sellAvailableFlag"`
		Currency              string     `json:"currency"`
		Sector                string     `json:"sector"`
		BuyAvailableFlag      bool       `json:"buyAvailableFlag"`
		ClassCode             string     `json:"classCode"`
		Ticker                string     `json:"ticker"`
		ApiTradeAvailableFlag bool       `json:"apiTradeAvailableFlag"`
		DlongMin              Quotation  `json:"dlongMin"`
		ShortEnabledFlag      bool       `json:"shortEnabledFlag"`
		Kshort                Quotation  `json:"kshort"`
		IssueSizePlan         string     `json:"issueSizePlan"`
		MinPriceIncrement     Quotation  `json:"minPriceIncrement"`
		OtcFlag               bool       `json:"otcFlag"`
		Klong                 Quotation  `json:"klong"`
		Dshort                Quotation  `json:"dshort"`
		Name                  string     `json:"name"`
		IssueSize             string     `json:"issueSize"`
		Exchange              string     `json:"exchange"`
		CountryOfRiskName     string     `json:"countryOfRiskName"`
		DivYieldFlag          bool       `json:"divYieldFlag"`
		Isin                  string     `json:"isin"`
		IpoDate               time.Time  `json:"ipoDate"`
	} `json:"instrument"`
}
