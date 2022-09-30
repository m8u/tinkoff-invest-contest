package db

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"log"
	"os"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

//var queryAPI api.QueryAPI
var writeAPI api.WriteAPI

func init() {
	url := "http://influxdb:8086"
	token := os.Getenv("INFLUXDB_TOKEN")
	org := "m8u"
	bucket := "tinkoff-invest-contest"

	client := influxdb2.NewClient(url, token)
	//queryAPI = client.QueryAPI(org)
	writeAPI = client.WriteAPI(org, bucket)

	err := client.DeleteAPI().DeleteWithName(context.Background(), org, bucket, time.Unix(0, 0), time.Now(), "")
	if err != nil {
		log.Fatalf("error: cannot empty the InfluxDB bucket (%v)", err.Error())
	}
}

func WriteStrategyOutput(botId int, strategyOutput map[string]any, ts time.Time) {
	writeAPI.WritePoint(write.NewPoint(
		fmt.Sprintf("bot_%v_strategy_output", botId),
		map[string]string{},
		strategyOutput,
		ts,
	))
}

func WriteHistoricCandles(botId int, candles []*investapi.HistoricCandle) {
	for _, candle := range candles {
		writeAPI.WritePoint(write.NewPoint(
			fmt.Sprintf("bot_%v_candles", botId),
			map[string]string{},
			marshalHistoricCandle(candle),
			candle.Time.AsTime(),
		))
	}
}

func WriteLastCandle(botId int, candle *investapi.Candle) {
	writeAPI.WritePoint(write.NewPoint(
		fmt.Sprintf("bot_%v_candles", botId),
		map[string]string{},
		marshalCandle(candle),
		candle.Time.AsTime(),
	))
}

func marshalHistoricCandle(candle *investapi.HistoricCandle) map[string]any {
	return map[string]any{
		"open":   utils.QuotationToFloat(candle.Open),
		"high":   utils.QuotationToFloat(candle.High),
		"low":    utils.QuotationToFloat(candle.Low),
		"close":  utils.QuotationToFloat(candle.Close),
		"volume": candle.Volume,
	}
}

func marshalCandle(candle *investapi.Candle) map[string]any {
	return map[string]any{
		"open":   utils.QuotationToFloat(candle.Open),
		"high":   utils.QuotationToFloat(candle.High),
		"low":    utils.QuotationToFloat(candle.Low),
		"close":  utils.QuotationToFloat(candle.Close),
		"volume": candle.Volume,
	}
}
