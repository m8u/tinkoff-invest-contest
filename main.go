package main

import (
	"flag"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const AppName = "m8u"

var ShouldExit = false
var NoInternetConnection = false

func main() {
	var mode = flag.String("mode", "",
		"Modes are:\n"+
			"'test'    - Run a test on historical data\n"+
			"'sandbox' - Start a sandbox bot\n"+
			"'combat'  - Start a combat bot",
	)
	var figi = flag.String("figi", "BBG006L8G4H1",
		"FIGI of a stock to trade",
	)
	var testDays = flag.Int("test_days", 100,
		"(for --mode=test) Number of days to test on",
	)
	var startMoney = flag.Float64("start_money", 100000,
		"(for --mode=test|sandbox) Starting money amount",
	)
	var fee = flag.Float64("fee", Fees[Premium],
		"(for --mode=test) Transaction fee (normalized, e.g. 0.00025 for 0.025%)",
	)
	var candleInterval = flag.String("candle_interval", "1min",
		"Candle interval. Possible values are:\n"+
			"'1min', "+
			//"'2min', "+
			//"'3min', "+
			"'5min', "+
			//"'10min', "+
			"'15min', "+
			//"'30min', "+
			"'hour'",
	)
	var window = flag.Int("window", 60,
		"Bollinger Bands MA window size",
	)
	var bollingerCoef = flag.Float64("bollinger_coef", 3,
		"Bollinger Bands coefficient (number of standard deviations)",
	)
	var maxPointDeviation = flag.Float64("max_point_dev", 0.001,
		"Maximum relative deviation when detecting price-bound intersections (normalized, e.g. 0.001 for 0.1%)",
	)
	var allowMargin = flag.Bool("allow_margin", false,
		"Either allow margin trading or not (1 or 0) (default: 0)",
	)
	flag.Parse()

	if _, ok := CandleIntervalsToDurations[sdk.CandleInterval(*candleInterval)]; !ok {
		log.Fatalln("please choose one of supported candle intervals\n" +
			"Try '--help' for more info")
	}

	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	_, err := os.ReadDir("logs")
	if err != nil {
		err = os.Mkdir("logs", 0775)
		MaybeCrash(err)
	}
	logFile, err := os.OpenFile("logs/"+time.Now().Format(time.RFC3339)+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("can't create log file: %v", err)
	}
	defer logFile.Close()
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	charts := Charts{
		Candles:        new([]sdk.Candle),
		Intervals:      new([][]float64),
		Flags:          new([][]ChartsTradeFlag),
		BalanceHistory: new([]float64),
		StartBalance:   new(float64),
		TestMode:       new(bool),
	}

	_ = godotenv.Load(".env")

	switch *mode {
	case "test":
		token := os.Getenv("SANDBOX_TOKEN")
		if token == "" {
			log.Fatalln("please provide sandbox token via 'SANDBOX_TOKEN' environment variable")
		}

		WaitForInternetConnection()

		*charts.TestMode = true
		log.Println("Testing on historical data...")
		TestOnHistoricalData(
			token,
			*figi,
			*testDays,
			sdk.CandleInterval(*candleInterval),
			StrategyParams{
				Window:                 *window,
				BollingerCoef:          *bollingerCoef,
				IntervalPointDeviation: *maxPointDeviation,
			},
			*startMoney,
			*fee,
			*allowMargin,
			&charts,
		)
		break
	case "sandbox":
		if *allowMargin {
			log.Fatalln("can't margin trade in sandbox")
		}
		token := os.Getenv("SANDBOX_TOKEN")
		if token == "" {
			log.Fatalln("please provide sandbox token via 'SANDBOX_TOKEN' environment variable")
		}

		WaitForInternetConnection()

		*charts.TestMode = false
		log.Println("Starting a sandbox bot...")
		bot := NewSandboxBot(
			token,
			*startMoney,
			*figi,
			sdk.CandleInterval(*candleInterval),
			*fee,
			StrategyParams{
				Window:                 *window,
				BollingerCoef:          *bollingerCoef,
				IntervalPointDeviation: *maxPointDeviation,
			},
			*allowMargin,
		)

		go bot.Serve(&charts)

		break
	case "combat":
		token := os.Getenv("COMBAT_TOKEN")
		if token == "" {
			log.Fatalln("please provide combat token via 'COMBAT_TOKEN' environment variable")
		}

		WaitForInternetConnection()

		*charts.TestMode = false
		log.Println("Starting a combat bot...")
		bot := NewCombatBot(
			token,
			*figi,
			sdk.CandleInterval(*candleInterval),
			StrategyParams{
				Window:                 *window,
				BollingerCoef:          *bollingerCoef,
				IntervalPointDeviation: *maxPointDeviation,
			},
			*allowMargin,
		)

		go bot.Serve(&charts)

		break
	default:
		log.Println("no mode specified\n" +
			"Try '--help' for more info")
		os.Exit(0)
	}

	if *mode != "test" {
		// запустим goroutine-у чтобы ждала один из сигналов прекращения работы
		// и устанавливала флаг ShouldExit
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func() {
			<-ch
			signal.Stop(ch)
			log.Println("Exiting...")
			ShouldExit = true
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}()
	}

	http.HandleFunc("/", charts.HandleTradingChart)
	http.HandleFunc("/balance", charts.HandleBalanceChart)
	log.Println("Starting echarts server at localhost:8081 (press Ctrl-C to exit)")
	log.Fatalln(http.ListenAndServe(":8081", nil))
}
