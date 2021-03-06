package bots

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
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

type TechnicalIndicatorBot struct {
	figi           string
	instrumentType utils.InstrumentType
	candleInterval investapi.CandleInterval
	window         int
	allowMargin    bool
	fee            float64

	id   string
	name string

	tradeEnv          *tradeenv.TradeEnv
	occupiedAccountId string

	strategy tistrategy.TechnicalIndicatorStrategy

	started  bool
	paused   bool
	removing bool
}

func NewTechnicalIndicatorBot(id string, name string, tradeEnv *tradeenv.TradeEnv, figi string, instrumentType utils.InstrumentType,
	candleInterval investapi.CandleInterval, window int, allowMargin bool, fee float64, strategy tistrategy.TechnicalIndicatorStrategy) *TechnicalIndicatorBot {
	bot := &TechnicalIndicatorBot{
		figi:           figi,
		instrumentType: instrumentType,
		candleInterval: candleInterval,
		window:         window,
		allowMargin:    allowMargin,
		fee:            fee,
		id:             id,
		name:           name,
		tradeEnv:       tradeEnv,
		strategy:       strategy,
	}

	bot.tradeEnv.InitChannels(bot.figi)

	err := db.CreateCandlesTable(bot.id)
	utils.MaybeCrash(err)

	db.CreateIndicatorValuesTable(bot.id, strategy.GetOutputKeys())

	err = dashboard.AddBotDashboard(bot.id, bot.name)
	utils.MaybeCrash(err)

	return bot
}

var prevDirection investapi.OrderDirection

func (bot *TechnicalIndicatorBot) loop() error {
	currentTimestamp := time.Time{}

	var candles []*investapi.HistoricCandle
	var err error

	instrument, err := bot.tradeEnv.Client.InstrumentByFigi(bot.figi, bot.instrumentType)
	if err != nil {
		log.Println(bot.logPrefix(), utils.PrettifyError(err))
		return err
	}

	for !appstate.ShouldExit && !bot.removing {
		select {
		// Get candle from stream
		case currentCandle := <-bot.tradeEnv.Channels[bot.figi].Candle:
			if currentCandle.Time.AsTime() != currentTimestamp {
				// On a new candle, get historic candles in amount of >= window
				candles, err = bot.tradeEnv.GetAtLeastNLastCandles(bot.figi, bot.candleInterval, bot.window)
				if err != nil {
					log.Println(bot.logPrefix(), utils.PrettifyError(err))
					return err
				}
				// Trim excessive candles
				candles = candles[len(candles)-(bot.window-1):]
				go func() {
					db.InsertCandles(bot.id, candles)
				}()
				currentTimestamp = currentCandle.Time.AsTime()
			}
			go func() {
				db.UpdateLastCandle(bot.id, currentCandle)
			}()

			// Get trade signal
			signal, indicatorValues := bot.strategy.GetTradeSignal(
				append(candles,
					&investapi.HistoricCandle{
						Open:   currentCandle.Open,
						High:   currentCandle.High,
						Low:    currentCandle.Low,
						Close:  currentCandle.Close,
						Volume: currentCandle.Volume,
					},
				),
			)
			indicatorValues["time"] = currentCandle.Time.AsTime()
			go func() {
				db.AddIndicatorValues(bot.id, indicatorValues)
			}()

			if signal != nil {
				// Get unoccupied account or use the existing one,
				// and determine lot quantity for the deal (either buy or sell)
				var lots int64
				shouldUnoccupyAccount := false
				if bot.occupiedAccountId == "" {
					accountId, unlock := bot.tradeEnv.GetUnoccupiedAccount(instrument.GetCurrency())
					if accountId == "" {
						unlock()
						continue
					}
					maxDealValue, err := bot.tradeEnv.CalculateMaxDealValue(
						accountId,
						signal.Direction,
						instrument,
						currentCandle.Close,
						bot.allowMargin,
					)
					if err != nil {
						log.Println(bot.logPrefix(), utils.PrettifyError(err))
						unlock()
						return err
					}
					lots = bot.tradeEnv.CalculateLotsCanAfford(signal.Direction, maxDealValue, instrument, currentCandle.Close, bot.fee)
					if lots == 0 {
						unlock()
						continue
					}
					bot.tradeEnv.SetAccountOccupied(accountId, instrument.GetCurrency())
					unlock()
					bot.occupiedAccountId = accountId
				} else if signal.Direction != prevDirection {
					shouldUnoccupyAccount = true
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
				log.Printf("%v %v %v lots for %v %v, account: %v",
					bot.logPrefix(),
					utils.OrderDirectionToString(signal.Direction),
					lots,
					utils.QuotationToFloat(currentCandle.Close),
					instrument.GetCurrency(),
					bot.occupiedAccountId,
				)
				err := bot.tradeEnv.DoOrder(
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
				dashboard.AnnotateOrder(
					bot.id,
					signal.Direction,
					lots*int64(instrument.GetLot()),
					utils.QuotationToFloat(currentCandle.Close),
					instrument.GetCurrency(),
				)

				if shouldUnoccupyAccount {
					bot.tradeEnv.SetAccountUnoccupied(bot.occupiedAccountId, instrument.GetCurrency())
					bot.occupiedAccountId = ""
				}

				prevDirection = signal.Direction
			}

		default:
			// Don't block, unless
			for bot.paused && !bot.removing {
			}
		}
	}
	return nil
}

func (bot *TechnicalIndicatorBot) Serve() {
	bot.started = true
	for !appstate.ShouldExit && !bot.removing {
		log.Printf("%v bot %q is starting...", bot.logPrefix(), bot.name)

		utils.WaitForInternetConnection()

		err := bot.tradeEnv.Client.SubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
		utils.MaybeCrash(err)

		err = bot.loop()
		if err != nil {
			log.Printf("%v bot %q has crashed, restarting...", bot.logPrefix(), bot.name)
			time.Sleep(10 * time.Second)
		}
	}
}

func (bot *TechnicalIndicatorBot) TogglePause() {
	bot.paused = !bot.paused
	if bot.paused {
		log.Printf("%v bot %q is paused", bot.logPrefix(), bot.name)
	} else {
		log.Printf("%v bot %q resumed, continue trading...", bot.logPrefix(), bot.name)
	}
}

func (bot *TechnicalIndicatorBot) Remove() {
	bot.removing = true
	err := bot.tradeEnv.Client.UnsubscribeCandles(bot.figi, investapi.SubscriptionInterval(bot.candleInterval))
	if err != nil {
		log.Println(bot.logPrefix(), utils.PrettifyError(err))
	}
	log.Printf("%v bot %q has been removed", bot.logPrefix(), bot.name)
}

func (bot *TechnicalIndicatorBot) IsPaused() bool {
	return bot.paused
}

func (bot *TechnicalIndicatorBot) IsStarted() bool {
	return bot.started
}

func (bot *TechnicalIndicatorBot) logPrefix() string {
	return fmt.Sprintf("[bot#%v]", bot.id)
}

func (bot *TechnicalIndicatorBot) GetYAML() string {
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
		Strategy:       bot.strategy.GetYAML(),
	}
	bytes, err := yaml.Marshal(obj)
	utils.MaybeCrash(err)
	return string(bytes)
}
