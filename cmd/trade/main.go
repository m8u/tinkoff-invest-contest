package main

import (
	"flag"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/metrics"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

func main() {
	var mode = flag.String("mode", "",
		"Modes are:\n"+
			"'sandbox' - Start a sandbox bot\n"+
			"'combat'  - Start a combat bot",
	)
	var figi = flag.String("figi", "BBG006L8G4H1",
		"FIGI of a stock to bollinger_bot",
	)
	var startMoney = flag.Float64("start_money", 100000,
		"(for --mode=sandbox) Starting money amount",
	)
	var fee = flag.Float64("fee", utils.Fees[utils.Premium],
		"(for --mode=sandbox) Transaction fee (normalized, e.g. 0.00025 for 0.025%)",
	)
	var candleInterval = flag.String("candle_interval", "1min",
		"Candle interval. Possible values are:\n"+
			"'1min' (realtime trading available)\n"+
			"'5min' (realtime trading available)\n"+
			"'15min'\n"+
			"'hour'\n",
	)
	//var window = flag.Int("window", 60,
	//	"Bollinger Bands MA window size",
	//)
	//var bollingerCoef = flag.Float64("bollinger_coef", 3,
	//	"Bollinger Bands coefficient (number of standard deviations)",
	//)
	//var maxPointDeviation = flag.Float64("max_point_dev", 0.001,
	//	"Maximum relative deviation when detecting price-bound intersections (normalized, e.g. 0.001 for 0.1%)",
	//)
	var allowMargin = flag.Bool("allow_margin", false,
		"(for --mode=combat) Either allow margin trading or not (1 or 0) (default: 0)",
	)
	flag.Parse()

	if _, ok := utils.CandleIntervalsToDurations[utils.CandleIntervalsV1NamesToValues[*candleInterval]]; !ok {
		log.Fatalln("please choose one of supported candle intervals\n" +
			"Try '--help' for more info")
	}

	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	_, err := os.ReadDir("logs")
	if err != nil {
		err = os.Mkdir("logs", 0775)
		utils.MaybeCrash(err)
	}
	logFile, err := os.OpenFile("logs/"+time.Now().Format(time.RFC3339)+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("can't create log file: %v", err)
	}
	defer logFile.Close()
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	charts := metrics.NewCharts()

	_ = godotenv.Load(".env")

	tradeEnvConfig := config.Config{
		IsSandbox:   true,
		Token:       os.Getenv("SANDBOX_TOKEN"),
		NumAccounts: 1,
		Money:       *startMoney,
		Fee:         *fee,
		Instruments: []config.ConfigInstrument{
			{
				FIGI:           *figi,
				CandleInterval: utils.CandleIntervalsV1NamesToValues[*candleInterval],
				OrderBookDepth: 10,
			},
		},
	}

	switch *mode {
	case "sandbox":
		if *allowMargin {
			log.Fatalln("can't margin-trade in sandbox")
		}
		token := os.Getenv("SANDBOX_TOKEN")
		if token == "" {
			log.Fatalln("please provide sandbox token via 'SANDBOX_TOKEN' environment variable")
		}

		utils.WaitForInternetConnection()

		log.Println("Starting a sandbox bot...")
		tradeEnv := tradeenv.New(tradeEnvConfig)
		bot := bots.New(tradeEnv, charts, *figi, utils.CandleIntervalsV1NamesToValues[*candleInterval])

		go bot.Serve()

		break
	case "combat":
		token := os.Getenv("COMBAT_TOKEN")
		if token == "" {
			log.Fatalln("please provide combat token via 'COMBAT_TOKEN' environment variable")
		}

		utils.WaitForInternetConnection()

		log.Println("Starting a combat bot...")
		tradeEnv := tradeenv.New(tradeEnvConfig)
		bot := bots.New(tradeEnv, charts, *figi, utils.CandleIntervalsV1NamesToValues[*candleInterval])

		go bot.Serve()

		break
	default:
		log.Println("no mode specified, or there is no such mode\n" +
			"Try '--help' for more info")
		os.Exit(0)
	}

	// запустим goroutine-у чтобы ждала один из сигналов прекращения работы
	// и устанавливала флаг ShouldExit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-ch
		signal.Stop(ch)
		log.Println("Exiting...")
		appstate.ShouldExit = true
		appstate.ExitChan <- true
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	// запускаем сервер с графиками
	http.HandleFunc("/", charts.HandleTradingChart)
	http.HandleFunc("/balance", charts.HandleBalanceChart)
	log.Println("Starting echarts server at localhost:8081 (press Ctrl-C to exit)")
	log.Fatalln(http.ListenAndServe(":8081", nil))
}
