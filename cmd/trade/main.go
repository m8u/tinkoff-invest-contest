package main

import (
	"flag"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/config"
	"tinkoff-invest-contest/internal/dashboard"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/strategies/ti/bollinger"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

func main() {
	var mode = flag.String("mode", "",
		"Modes are:\n"+
			"'sandbox' - Start a sandbox bot\n"+
			"'combat'  - Start a combat bot",
	)
	var figi = flag.String("figi", "BBG004730RP0",
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

	_ = godotenv.Load(".env")

	tradeEnvConfig := config.Config{
		IsSandbox:   true,
		Token:       utils.GetSandboxToken(),
		NumAccounts: 1,
		Money:       *startMoney,
		Fee:         *fee,
	}

	db.InitDB()
	dashboard.InitGrafana()

	switch *mode {
	case "sandbox":
		if *allowMargin {
			log.Fatalln("can't margin-trade in sandbox")
		}

		utils.WaitForInternetConnection()

		log.Println("Creating sandbox trade environment...")
		tradeEnv := tradeenv.New(tradeEnvConfig)
		bot := bots.New(tradeEnv, *figi, utils.CandleIntervalsV1NamesToValues[*candleInterval], *window,
			bollinger.New(*bollingerCoef, *maxPointDeviation))

		go bot.Serve()

		break
	case "combat":
		utils.WaitForInternetConnection()

		log.Println("Creating combat trade environment...")
		tradeEnv := tradeenv.New(tradeEnvConfig)
		bot := bots.New(tradeEnv, *figi, utils.CandleIntervalsV1NamesToValues[*candleInterval], *window,
			bollinger.New(*bollingerCoef, *maxPointDeviation))

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
	func() {
		<-ch
		signal.Stop(ch)
		log.Println("Exiting...")
		appstate.ShouldExit = true
		appstate.ExitChan <- true

		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()
}
