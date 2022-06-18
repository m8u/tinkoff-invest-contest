package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
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

func InsertCandles(figi string, candles []*investapi.HistoricCandle) error {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO %v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time) ON CONFLICT DO NOTHING`, figi), sqlizeHistoricCandles(candles))
	if err != nil {
		return err
	}
	return nil
}

func UpdateLastCandle(figi string, candle *investapi.Candle) error {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`UPDATE %v_candles
		SET open=:open, high=:high, low=:low, close=:close, volume=:volume
		WHERE time=:time`, figi), sqlizeCandle(candle))
	if err != nil {
		return err
	}
	return nil
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
			utils.FloatFromQuotation(candle.Open),
			utils.FloatFromQuotation(candle.High),
			utils.FloatFromQuotation(candle.Low),
			utils.FloatFromQuotation(candle.Close),
			candle.Volume,
			candle.Time.AsTime(),
		})
	}
	return sqlizedCandles
}

func sqlizeCandle(candle *investapi.Candle) any {
	return sqlCandle{
		utils.FloatFromQuotation(candle.Open),
		utils.FloatFromQuotation(candle.High),
		utils.FloatFromQuotation(candle.Low),
		utils.FloatFromQuotation(candle.Close),
		candle.Volume,
		candle.Time.AsTime(),
	}
}
