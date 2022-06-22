package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"time"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

const Host = "db"
const User = "postgres"
const Password = "postgres"
const DBname = "tinkoff_invest_contest"

var db *sqlx.DB

func InitDB() {
	connStr := fmt.Sprintf("host=%v user=%v password=%v dbname=%v sslmode=disable",
		Host, User, Password, DBname)

	var err error
	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
}

func ensureDBInitialized() {
	if db == nil {
		log.Fatalln("database was not initialized")
	}
}

func CreateCandlesTable(figi string) error {
	ensureDBInitialized()
	_, err := db.Exec(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v_candles (
			open DOUBLE PRECISION,
			high DOUBLE PRECISION,
			low DOUBLE PRECISION,
			close DOUBLE PRECISION,
			volume BIGINT,
			time TIMESTAMP WITH TIME ZONE UNIQUE)`,
			figi,
		),
	)
	if err != nil {
		return err
	}
	return nil
}

func CreateIndicatorValuesTable(figi string, descriptor []string) {
	ensureDBInitialized()
	sqlStr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v_indicators (", figi)
	for _, name := range descriptor {
		sqlStr += name + " DOUBLE PRECISION, "
	}
	sqlStr += "time TIMESTAMP WITH TIME ZONE UNIQUE)"
	_, err := db.Exec(sqlStr)
	utils.MaybeCrash(err)
}

func AddIndicatorValues(figi string, indicatorValues map[string]any) {
	ensureDBInitialized()
	keys := make([]string, 0)
	values := make([]any, 0)
	sqlStr := fmt.Sprintf("INSERT INTO %v_indicators(", figi)
	for k, v := range indicatorValues {
		keys = append(keys, k)
		values = append(values, v)
	}
	sqlStr += strings.Join(keys, ", ")
	sqlStr += ") VALUES (:"
	sqlStr += strings.Join(keys, ", :")
	sqlStr += ") ON CONFLICT (time) DO UPDATE SET "
	for _, key := range keys {
		if key == "time" {
			continue
		}
		sqlStr += key + "=excluded." + key + ","
	}
	sqlStr = sqlStr[:len(sqlStr)-1]
	_, err := db.NamedExec(sqlStr, indicatorValues)
	utils.MaybeCrash(err)
}

func InsertCandles(figi string, candles []*investapi.HistoricCandle) {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO %v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time)
		ON CONFLICT (time) DO UPDATE
		SET open=excluded.open, high=excluded.high, low=excluded.low, close=excluded.close, volume=excluded.volume`,
		figi), sqlizeHistoricCandles(candles))
	utils.MaybeCrash(err)
}

func UpdateLastCandle(figi string, candle *investapi.Candle) {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO %v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time)
		ON CONFLICT (time) DO UPDATE
		SET open=excluded.open, high=excluded.high, low=excluded.low, close=excluded.close, volume=excluded.volume`,
		figi), sqlizeCandle(candle))
	utils.MaybeCrash(err)
}

type sqlCandle struct {
	Open   float64   `db:"open"`
	High   float64   `db:"high"`
	Low    float64   `db:"low"`
	Close  float64   `db:"close"`
	Volume int64     `db:"volume"`
	Time   time.Time `db:"time"`
}

func sqlizeHistoricCandles(candles []*investapi.HistoricCandle) []any {
	sqlizedCandles := make([]any, 0)
	for _, candle := range candles {
		sqlizedCandles = append(sqlizedCandles, sqlCandle{
			Open:   utils.FloatFromQuotation(candle.Open),
			High:   utils.FloatFromQuotation(candle.High),
			Low:    utils.FloatFromQuotation(candle.Low),
			Close:  utils.FloatFromQuotation(candle.Close),
			Volume: candle.Volume,
			Time:   candle.Time.AsTime(),
		})
	}
	return sqlizedCandles
}

func sqlizeCandle(candle *investapi.Candle) any {
	return sqlCandle{
		Open:   utils.FloatFromQuotation(candle.Open),
		High:   utils.FloatFromQuotation(candle.High),
		Low:    utils.FloatFromQuotation(candle.Low),
		Close:  utils.FloatFromQuotation(candle.Close),
		Volume: candle.Volume,
		Time:   candle.Time.AsTime(),
	}
}