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

const connStr string = "host=db user=postgres password=postgres dbname=tinkoff_invest_contest sslmode=disable"

var db *sqlx.DB

func InitDB() {
	var err error
	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}
}

func assureDBInitialized() {
	if db == nil {
		log.Fatalln("database was not initialized")
	}
}

func CreateCandlesTable(figi string) error {
	assureDBInitialized()
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
	assureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO %v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time) ON CONFLICT DO NOTHING`, figi), batchSQLize(candles))
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

func batchSQLize(candles []*investapi.HistoricCandle) []any {
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
