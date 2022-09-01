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

func init() {
	connStr := fmt.Sprintf("host=%v user=%v password=%v dbname=%v sslmode=disable",
		Host, User, Password, DBname)

	var err error
	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	_, err = db.Exec("DROP SCHEMA public CASCADE")
	_, _ = db.Exec("CREATE SCHEMA public")
	_, _ = db.Exec("GRANT ALL ON SCHEMA public TO postgres")
	_, _ = db.Exec("GRANT ALL ON SCHEMA public TO public")
	utils.MaybeCrash(err)
}

func ensureDBInitialized() {
	if db == nil {
		log.Fatalln("database connection was not initialized")
	}
}

func CreateCandlesTable(botId string) {
	ensureDBInitialized()
	_, err := db.Exec(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS bot_%v_candles (
			open DOUBLE PRECISION,
			high DOUBLE PRECISION,
			low DOUBLE PRECISION,
			close DOUBLE PRECISION,
			volume BIGINT,
			time TIMESTAMP WITH TIME ZONE UNIQUE)`,
			botId,
		),
	)
	utils.MaybeCrash(err)
}

func CreateIndicatorValuesTable(botId string, fields []string) {
	ensureDBInitialized()
	sqlStr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS bot_%v_indicators (", botId)
	for _, name := range fields {
		sqlStr += name + " DOUBLE PRECISION, "
	}
	sqlStr += "time TIMESTAMP WITH TIME ZONE UNIQUE)"
	_, err := db.Exec(sqlStr)
	utils.MaybeCrash(err)
}

func AddStrategyOutputValues(botId string, indicatorValues map[string]any) {
	ensureDBInitialized()
	keys := make([]string, 0)
	values := make([]any, 0)
	sqlStr := fmt.Sprintf("INSERT INTO bot_%v_indicators(", botId)
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

func InsertCandles(botId string, candles []*investapi.HistoricCandle) {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO bot_%v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time)
		ON CONFLICT (time) DO UPDATE
		SET open=excluded.open, high=excluded.high, low=excluded.low, close=excluded.close, volume=excluded.volume`,
		botId), sqlizeHistoricCandles(candles))
	utils.MaybeCrash(err)
}

func UpdateLastCandle(botId string, candle *investapi.Candle) {
	ensureDBInitialized()
	_, err := db.NamedExec(fmt.Sprintf(`INSERT INTO bot_%v_candles(open, high, low, close, volume, time)
		VALUES (:open, :high, :low, :close, :volume, :time)
		ON CONFLICT (time) DO UPDATE
		SET open=excluded.open, high=excluded.high, low=excluded.low, close=excluded.close, volume=excluded.volume`,
		botId), sqlizeCandle(candle))
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
			Open:   utils.QuotationToFloat(candle.Open),
			High:   utils.QuotationToFloat(candle.High),
			Low:    utils.QuotationToFloat(candle.Low),
			Close:  utils.QuotationToFloat(candle.Close),
			Volume: candle.Volume,
			Time:   candle.Time.AsTime(),
		})
	}
	return sqlizedCandles
}

func sqlizeCandle(candle *investapi.Candle) any {
	return sqlCandle{
		Open:   utils.QuotationToFloat(candle.Open),
		High:   utils.QuotationToFloat(candle.High),
		Low:    utils.QuotationToFloat(candle.Low),
		Close:  utils.QuotationToFloat(candle.Close),
		Volume: candle.Volume,
		Time:   candle.Time.AsTime(),
	}
}
