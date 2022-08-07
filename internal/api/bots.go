package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
	"tinkoff-invest-contest/internal/bot"
	"tinkoff-invest-contest/internal/strategies/strategy"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

var botId int64

func CreateBot(c *gin.Context) {
	args := struct {
		Sandbox        bool    `form:"sandbox"`
		Figi           string  `form:"figi"`
		InstrumentType string  `form:"instrumentType"`
		AllowMargin    bool    `form:"allowMargin"`
		Fee            float64 `form:"fee"`

		StrategyName   string `form:"strategyName"`
		StrategyConfig string `form:"strategyConfig"`
		CandleInterval string `form:"candleInterval"`
		Window         int    `form:"window"`
		OrderBookDepth int32  `form:"orderBookDepth"`
	}{}

	err := c.Bind(&args)
	if err != nil {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusBadRequest,
			"One or more arguments are invalid ("+err.Error()+")",
		))
		return
	}

	var tradeEnv *tradeenv.TradeEnv
	var fee float64
	if args.Sandbox {
		tradeEnv = app.SandboxEnv
		fee = args.Fee / 100
	} else {
		tradeEnv = app.CombatEnv
		fee = tradeEnv.CombatFee
	}
	instrumentType, err := utils.StringToInstrumentType(args.InstrumentType)
	if err != nil {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusBadRequest,
			"No such instrument type '"+args.InstrumentType+"' ("+err.Error()+")",
		))
		return
	}

	id := fmt.Sprint(botId)
	instrument, err := tradeEnv.Client.InstrumentByFigi(args.Figi, instrumentType)
	if err != nil {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusNotFound,
			"Couldn't find instrument by FIGI '"+args.Figi+"' ("+err.Error()+")",
		))
		return
	}
	name := instrument.GetTicker()
	if args.Sandbox {
		name = "[sandbox] " + name
	}
	name += " #" + id

	if newStrategyFromJSON, ok := strategy.JSONConstructors[args.StrategyName]; ok {
		s, err := newStrategyFromJSON(args.StrategyConfig)
		if err != nil {
			_, _ = c.Writer.WriteString(marshalResponse(
				http.StatusBadRequest,
				"Invalid strategy config for '"+args.StrategyName+"' ("+err.Error()+")",
			))
			return
		}
		candleInterval, err := utils.StringToCandleInterval(args.CandleInterval)
		if err != nil {
			_, _ = c.Writer.WriteString(marshalResponse(
				http.StatusBadRequest,
				"Invalid candle interval '"+args.CandleInterval+"' ("+err.Error()+")",
			))
			return
		}

		app.Bots.Lock.Lock()
		app.Bots.Table[id] = bot.NewTechnicalIndicatorBot(id, name, tradeEnv, args.Figi, instrumentType, candleInterval, args.Window, args.OrderBookDepth, args.AllowMargin, fee, s)
		app.Bots.Lock.Unlock()
	} else {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusBadRequest,
			"No known strategy specified",
		))
		return
	}

	botId++
	_, _ = c.Writer.WriteString(marshalResponse(
		http.StatusOK,
		"",
		struct {
			Name string `json:"name"`
		}{name},
	))
}

func StartBot(c *gin.Context) {
	id := c.Query("id")
	go app.Bots.Table[id].Serve()

	_, _ = c.Writer.WriteString("ok")
}

func TogglePauseBot(c *gin.Context) {
	id := c.Query("id")
	app.Bots.Table[id].TogglePause()

	_, _ = c.Writer.WriteString("ok")
}

func RemoveBot(c *gin.Context) {
	id := c.Query("id")
	if app.Bots.Table[id].IsPaused() {
		app.Bots.Table[id].TogglePause()
	}
	app.Bots.Table[id].Remove()
	app.Bots.Lock.Lock()
	delete(app.Bots.Table, id)
	app.Bots.Lock.Unlock()

	_, _ = c.Writer.WriteString("ok")
}
