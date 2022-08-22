package bot

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"log"
	"math"
	"time"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/dashboard"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/strategies"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type Bot struct {
	figi           string
	instrumentType utils.InstrumentType
	candleInterval investapi.CandleInterval
	window         int
	orderBookDepth int32
	allowMargin    bool
	fee            float64

	id   string
	name string

	tradeEnv            *tradeenv.TradeEnv
	occupiedAccountId   string
	lastDiscardTS       time.Time
	prevSignalDirection investapi.OrderDirection

	strategy strategies.Strategy

	started  bool
	paused   bool
	removing bool
}

func New(id string, name string, tradeEnv *tradeenv.TradeEnv, figi string,
	instrumentType utils.InstrumentType, candleInterval investapi.CandleInterval, window int, orderBookDepth int32,
	allowMargin bool, fee float64, strategy strategies.Strategy) *Bot {
	bot := &Bot{
		figi:           figi,
		instrumentType: instrumentType,
		candleInterval: candleInterval,
		window:         window,
		orderBookDepth: orderBookDepth,
		allowMargin:    allowMargin,
		fee:            fee,
		id:             id,
		name:           name,
		tradeEnv:       tradeEnv,
		strategy:       strategy,
	}

	bot.tradeEnv.InitMarketDataChannels(bot.figi)

	err := db.CreateCandlesTable(bot.id)
	utils.MaybeCrash(err)

	db.CreateIndicatorValuesTable(bot.id, strategy.GetOutputKeys())

	err = dashboard.AddBotDashboard(bot.id, bot.name)
	utils.MaybeCrash(err)

	return bot
}

func (bot *Bot) loop() error {
	currentTimestamp := time.Time{}

	var (
		candles              []*investapi.HistoricCandle
		err                  error
		currentCandle        *investapi.Candle
		currentOrderBook     *investapi.OrderBook
		shouldReleaseAccount bool
	)

	instrument, err := bot.tradeEnv.Client.InstrumentByFigi(bot.figi, bot.instrumentType)
	if err != nil {
		log.Println(bot.logPrefix(), utils.PrettifyError(err))
		return err
	}

	for !appstate.ShouldExit && !bot.removing {
		marketData := bot.tradeEnv.GetMarketDataChannels(bot.figi)
		select {
		// Get candle from stream
		case candle := <-marketData.Candle:
			currentCandle = candle
			if currentCandle.Time.AsTime() != currentTimestamp {
				// On a new candle, get historic candles in amount of >= window
				candles, err = bot.tradeEnv.GetAtLeastNLastCandles(bot.figi, bot.candleInterval, bot.window)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
					return err
				}
				// Trim excessive candles
				candles = candles[len(candles)-(bot.window-1):]
				db.InsertCandles(bot.id, candles)
				currentTimestamp = currentCandle.Time.AsTime()
			}
			go db.UpdateLastCandle(bot.id, currentCandle)

		case orderBook := <-marketData.OrderBook:
			currentOrderBook = orderBook

		default:
			// Don't block, unless
			for bot.paused && !bot.removing {
			}
			continue
		}

		if currentCandle == nil || currentOrderBook == nil {
			continue
		}

		// Get trade signal
		signal, outputValues := bot.strategy.GetTradeSignal(
			strategies.MarketData{
				Candles: append(candles,
					&investapi.HistoricCandle{
						Open:   currentCandle.Open,
						High:   currentCandle.High,
						Low:    currentCandle.Low,
						Close:  currentCandle.Close,
						Volume: currentCandle.Volume,
					},
				),
				OrderBook: currentOrderBook,
			},
		)
		if len(outputValues) > 0 {
			outputValues["time"] = currentCandle.Time.AsTime()
			go db.AddStrategyOutputValues(bot.id, outputValues)
		}

		shouldReleaseAccount = false
		if signal != nil && time.Now().After(bot.lastDiscardTS.Add(time.Minute)) {
			// Get unoccupied account or use the existing one,
			// and determine lot quantity for the deal (either buy or sell)
			var lots int64
			if bot.occupiedAccountId == "" {
				accountId, discard, unlock := bot.tradeEnv.GetUnoccupiedAccount(instrument.GetCurrency())
				if accountId == "" {
					continue
				}
				maxDealValue := bot.tradeEnv.CalculateMaxDealValue(
					accountId,
					signal.Direction,
					instrument,
					currentCandle.Close,
					bot.allowMargin,
				)
				lots = bot.tradeEnv.CalculateLotsCanAfford(signal.Direction, maxDealValue, instrument, currentCandle.Close, bot.fee)
				if lots == 0 {
					bot.lastDiscardTS = time.Now()
					discard()
					unlock()
					continue
				}
				unlock()
				bot.occupiedAccountId = accountId
			} else if signal.Direction != bot.prevSignalDirection {
				shouldReleaseAccount = true
				lots, err = bot.tradeEnv.GetLotsHave(bot.occupiedAccountId, instrument)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
					return err
				}
				lots = int64(math.Abs(float64(lots)))
			} else {
				continue
			}

			// Place an order and wait for it to be filled
			avgPositionPrice, err := bot.tradeEnv.DoOrder(
				bot.figi,
				lots,
				currentCandle.Close,
				signal.Direction,
				bot.occupiedAccountId,
				investapi.OrderType_ORDER_TYPE_MARKET,
			)
			if err != nil {
				log.Printf("%v order error: %v", bot.logPrefix(), utils.PrettifyError(err))
				return err
			}
			log.Printf("%v %v %v %v for avg. %v %v, account: %v",
				bot.logPrefix(),
				utils.OrderDirectionToString(signal.Direction),
				lots*int64(instrument.GetLot()),
				instrument.GetTicker(),
				avgPositionPrice,
				instrument.GetCurrency(),
				bot.occupiedAccountId,
			)
			err = dashboard.AnnotateOrder(
				bot.id,
				signal.Direction,
				lots*int64(instrument.GetLot()),
				avgPositionPrice,
				instrument.GetCurrency(),
			)
			if err != nil {
				log.Println(bot.logPrefix(), utils.PrettifyError(err))
			}

			if shouldReleaseAccount {
				bot.tradeEnv.ReleaseAccount(bot.occupiedAccountId, instrument.GetCurrency())
				bot.occupiedAccountId = ""
			}
			bot.prevSignalDirection = signal.Direction
		}
	}
	return nil
}

func (bot *Bot) Serve() {
	bot.started = true
	for !appstate.ShouldExit && !bot.removing {
		log.Printf("%v bot %q is starting...", bot.logPrefix(), bot.name)

		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)
		err = bot.tradeEnv.Client.SubscribeOrderBook(bot.figi, bot.orderBookDepth)
		utils.MaybeCrash(err)

		err = bot.loop()
		if err != nil {
			log.Printf("%v bot %q has crashed, restarting...", bot.logPrefix(), bot.name)
			time.Sleep(10 * time.Second)
		}
	}
}

func (bot *Bot) TogglePause() {
	bot.paused = !bot.paused
	if bot.paused {
		log.Printf("%v bot %q is paused", bot.logPrefix(), bot.name)
	} else {
		log.Printf("%v bot %q resumed, continue trading...", bot.logPrefix(), bot.name)
	}
}

func (bot *Bot) Remove() {
	bot.removing = true
	err := bot.tradeEnv.Client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	if err != nil {
		log.Println(bot.logPrefix(), utils.PrettifyError(err))
	}
	log.Printf("%v bot %q has been removed", bot.logPrefix(), bot.name)
}

func (bot *Bot) IsPaused() bool {
	return bot.paused
}

func (bot *Bot) IsStarted() bool {
	return bot.started
}

func (bot *Bot) logPrefix() string {
	return fmt.Sprintf("[bot#%v]", bot.id)
}

func (bot *Bot) GetYAML() string {
	obj := struct {
		FIGI        string  `yaml:"FIGI"`
		AllowMargin bool    `yaml:"AllowMargin"`
		Fee         float64 `yaml:"Fee"`

		Window         int    `yaml:"Window"`
		CandleInterval string `yaml:"CandleInterval"`

		Strategy any `yaml:"Strategy"`
	}{
		FIGI:           bot.figi,
		AllowMargin:    bot.allowMargin,
		Fee:            bot.fee,
		Window:         bot.window,
		CandleInterval: utils.CandleIntervalToString(bot.candleInterval),
		Strategy: struct {
			Name   string `yaml:"Name"`
			Params any    `yaml:"Params"`
		}{
			Name:   bot.strategy.GetName(),
			Params: bot.strategy.GetYAML(),
		},
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}
