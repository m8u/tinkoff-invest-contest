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
	id   int
	name string

	instrument  utils.InstrumentInterface
	allowMargin bool
	fee         float64

	tradeEnv *tradeenv.TradeEnv

	occupiedAccountId   string
	lastDiscardTS       time.Time
	prevSignalDirection investapi.OrderDirection

	ordersConfig strategies.OrdersConfig

	candleInterval investapi.CandleInterval
	window         int
	orderBookDepth int32
	strategy       strategies.Strategy

	started, paused, removing bool
	waitingForOrderExecution  bool
	orderError                chan error

	currentStopLoss, currentTakeProfit *strategies.TradeSignalStopOrder
}

func New(
	id int,
	name string,
	instrument utils.InstrumentInterface,
	allowMargin bool,
	fee float64,
	tradeEnv *tradeenv.TradeEnv,
	orderType investapi.OrderType,
	stopLossOrderType investapi.OrderType,
	takeProfitRatio float64,
	stopLossRatio float64,
	stopLossExecRatio float64,
	candleInterval investapi.CandleInterval,
	window int,
	orderBookDepth int32,
	strategy strategies.Strategy,
) *Bot {
	bot := &Bot{
		id:          id,
		name:        name,
		instrument:  instrument,
		allowMargin: allowMargin,
		fee:         fee,
		tradeEnv:    tradeEnv,
		ordersConfig: strategies.OrdersConfig{
			OrderType:         orderType,
			StopLossOrderType: stopLossOrderType,
			TakeProfitRatio:   takeProfitRatio,
			StopLossRatio:     stopLossRatio,
			StopLossExecRatio: stopLossExecRatio,
		},
		candleInterval: candleInterval,
		window:         window,
		orderBookDepth: orderBookDepth,
		strategy:       strategy,
		orderError:     make(chan error),
	}

	bot.tradeEnv.InitNewMarketDataChannels()

	dashboard.AddBotDashboard(bot.id, bot.name)

	return bot
}

func (bot *Bot) loop() error {
	log.Printf("%v bot %q has started", bot.logPrefix(), bot.name)
	currentTimestamp := time.Time{}
	var (
		candles              []*investapi.HistoricCandle
		err                  error
		currentCandle        *investapi.Candle
		currentOrderBook     *investapi.OrderBook
		shouldReleaseAccount bool
	)
	marketData := bot.tradeEnv.GetMarketDataChannels(bot.id)
	for !appstate.ShouldExit && !bot.removing {
		select {
		// Get candle from stream
		case candle := <-marketData.Candle:
			currentCandle = candle
			if currentCandle.Time.AsTime() != currentTimestamp {
				// On a new candle, get historic candles in amount of >= window
				candles, err = bot.tradeEnv.GetAtLeastNLastCandles(bot.instrument.GetFigi(), bot.candleInterval, bot.window)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
					return err
				}
				// Trim excessive candles
				candles = candles[len(candles)-(bot.window-1):]
				db.WriteHistoricCandles(bot.id, candles)
				currentTimestamp = currentCandle.Time.AsTime()
			}
			go db.WriteLastCandle(bot.id, currentCandle)

		case orderBook := <-marketData.OrderBook:
			if !orderBook.IsConsistent || len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
				continue
			}
			currentOrderBook = orderBook

		case orderError := <-bot.orderError:
			if orderError != nil {
				log.Printf("%v order error: %v", bot.logPrefix(), utils.PrettifyError(orderError))
				return orderError
			}
			bot.waitingForOrderExecution = false

		default:
			for bot.paused && !bot.removing {
				time.Sleep(2 * time.Second)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if currentCandle == nil || currentOrderBook == nil {
			continue
		}

		// Get trade signal
		signal, outputValues := bot.strategy.GetTradeSignal(
			bot.instrument,
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
			bot.ordersConfig,
		)
		if len(outputValues) > 0 {
			go db.WriteStrategyOutput(bot.id, outputValues, currentCandle.Time.AsTime())
		}

		if bot.waitingForOrderExecution {
			continue
		}

		if bot.currentStopLoss != nil {
			signal = nil
			if bot.currentStopLoss.IsTriggered(currentCandle.Close) {
				signal = &strategies.TradeSignal{
					Order: &strategies.TradeSignalOrder{
						Direction: bot.currentStopLoss.Direction,
					},
				}
				if bot.currentStopLoss.Type == investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LIMIT {
					signal.Order.Type = investapi.OrderType_ORDER_TYPE_LIMIT
					signal.Order.Price = bot.currentStopLoss.ExecPrice
				} else {
					signal.Order.Type = investapi.OrderType_ORDER_TYPE_MARKET
					signal.Order.Price = bot.currentStopLoss.TriggerPrice
				}
			} else if bot.currentTakeProfit.IsTriggered(currentCandle.Close) {
				signal = &strategies.TradeSignal{
					Order: &strategies.TradeSignalOrder{
						Type:      investapi.OrderType_ORDER_TYPE_MARKET,
						Direction: bot.currentTakeProfit.Direction,
						Price:     bot.currentTakeProfit.TriggerPrice,
					},
				}
			}
			if signal != nil {
				bot.currentStopLoss, bot.currentTakeProfit = nil, nil
			}
		}

		shouldReleaseAccount = false
		if signal != nil && time.Now().After(bot.lastDiscardTS.Add(time.Minute)) {
			// Get unoccupied account or use the existing one,
			// and determine lot quantity for the deal (either buy or sell)
			var lots int64
			if bot.occupiedAccountId == "" {
				accountId, discard, unlock := bot.tradeEnv.GetUnoccupiedAccount(bot.instrument.GetCurrency())
				if accountId == "" {
					continue
				}
				maxDealValue := bot.tradeEnv.CalculateMaxDealValue(
					accountId,
					signal.Order.Direction,
					bot.instrument,
					currentCandle.Close,
					bot.allowMargin,
				)
				lots = bot.tradeEnv.CalculateLotsCanAfford(signal.Order.Direction, maxDealValue, bot.instrument, currentCandle.Close, bot.fee)
				if lots == 0 {
					bot.lastDiscardTS = time.Now()
					discard()
					unlock()
					continue
				}
				unlock()
				bot.occupiedAccountId = accountId
			} else if signal.Order.Direction != bot.prevSignalDirection {
				shouldReleaseAccount = true
				lots, err = bot.tradeEnv.GetLotsHave(bot.occupiedAccountId, bot.instrument)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
					return err
				}
				lots = int64(math.Abs(float64(lots)))
			} else {
				continue
			}

			bot.waitingForOrderExecution = true
			go func() {
				// Place an order and wait for it to be filled
				avgPositionPrice, err := bot.tradeEnv.DoOrder(
					bot.instrument.GetFigi(),
					lots,
					signal.Order.Price,
					signal.Order.Direction,
					bot.occupiedAccountId,
					signal.Order.Type,
				)
				log.Println(bot.logPrefix())
				log.Printf("%v %v %v %v for %v %v (actual avg. %v %v), account: %v",
					bot.logPrefix(),
					utils.OrderDirectionToString(signal.Order.Direction),
					lots*int64(bot.instrument.GetLot()),
					bot.instrument.GetTicker(),
					utils.QuotationToFloat(signal.Order.Price),
					bot.instrument.GetCurrency(),
					avgPositionPrice,
					bot.instrument.GetCurrency(),
					bot.occupiedAccountId,
				)
				err = dashboard.AnnotateOrder(
					bot.id,
					signal.Order.Direction,
					lots*int64(bot.instrument.GetLot()),
					avgPositionPrice,
					bot.instrument.GetCurrency(),
				)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
				}

				if shouldReleaseAccount {
					bot.tradeEnv.ReleaseAccount(bot.occupiedAccountId, bot.instrument.GetCurrency())
					bot.occupiedAccountId = ""
				}
				bot.prevSignalDirection = signal.Order.Direction

				if signal.StopLoss != nil {
					bot.currentStopLoss, bot.currentTakeProfit = signal.StopLoss, signal.TakeProfit
					if bot.currentStopLoss.Type == investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LIMIT {
						log.Printf("%v setting stop loss = %v -> %v %v",
							bot.logPrefix(),
							utils.QuotationToFloat(bot.currentStopLoss.TriggerPrice),
							utils.QuotationToFloat(bot.currentStopLoss.ExecPrice),
							bot.instrument.GetCurrency(),
						)
					} else {
						log.Printf("%v setting stop loss = %v %v",
							bot.logPrefix(),
							utils.QuotationToFloat(bot.currentStopLoss.TriggerPrice),
							bot.instrument.GetCurrency(),
						)
					}
					log.Printf("%v setting take profit = %v %v",
						bot.logPrefix(),
						utils.QuotationToFloat(bot.currentTakeProfit.TriggerPrice),
						bot.instrument.GetCurrency(),
					)
				}
				log.Println(bot.logPrefix())

				bot.orderError <- err
			}()
		}
	}
	return nil
}

func (bot *Bot) Serve() {
	bot.started = true
	for !appstate.ShouldExit && !bot.removing {
		utils.WaitForInternetConnection()
		bot.tradeEnv.SubscribeCandles(bot.id, bot.instrument.GetFigi(), investapi.SubscriptionInterval(bot.candleInterval))
		bot.tradeEnv.SubscribeOrderBook(bot.id, bot.instrument.GetFigi(), bot.orderBookDepth)

		err := bot.loop()
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
	bot.tradeEnv.UnsubscribeAll(bot.id)
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
		FIGI:           bot.instrument.GetFigi(),
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
