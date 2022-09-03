package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"tinkoff-invest-contest/internal/app"
	"tinkoff-invest-contest/internal/bot"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/strategies"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

var mu sync.Mutex
var botId int

func CreateBot(c *gin.Context) {
	args := struct {
		Sandbox        bool                 `form:"sandbox"`
		Figi           string               `form:"figi"`
		InstrumentType utils.InstrumentType `form:"instrumentType"`
		AllowMargin    bool                 `form:"allowMargin"`

		StrategyName   string `form:"strategyName"`
		StrategyConfig string `form:"strategyConfig"`

		CandleInterval investapi.CandleInterval `form:"candleInterval"`
		Window         int                      `form:"window"`
		OrderBookDepth int32                    `form:"orderBookDepth"`

		OrderType         investapi.OrderType `form:"orderType"`
		StopLossOrderType investapi.OrderType `form:"stopLossOrderType"`
		TakeProfitRatio   float64             `form:"takeProfitRatio"`
		StopLossRatio     float64             `form:"stopLossRatio"`
		StopLossExecRatio float64             `form:"stopLossExecRatio"`
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
	if args.Sandbox {
		tradeEnv = app.SandboxEnv
	} else {
		tradeEnv = app.CombatEnv
	}
	instrument, err := tradeEnv.Client.InstrumentByFigi(args.Figi, args.InstrumentType)
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
	mu.Lock()
	name += " #" + fmt.Sprint(botId)

	if newStrategyFromJSON, ok := strategies.JSONConstructors[args.StrategyName]; ok {
		strategy, err := newStrategyFromJSON(args.StrategyConfig)
		if err != nil {
			_, _ = c.Writer.WriteString(marshalResponse(
				http.StatusBadRequest,
				"Invalid strategy config for '"+args.StrategyName+"' ("+err.Error()+")",
			))
			return
		}
		app.Bots.Table[fmt.Sprint(botId)] = bot.New(
			botId,
			name,
			instrument,
			args.AllowMargin,
			tradeEnv.Fee,
			tradeEnv,
			args.OrderType,
			args.StopLossOrderType,
			args.TakeProfitRatio,
			args.StopLossRatio,
			args.StopLossExecRatio,
			args.CandleInterval,
			args.Window,
			args.OrderBookDepth,
			strategy,
		)
	} else {
		_, _ = c.Writer.WriteString(marshalResponse(
			http.StatusBadRequest,
			"No known strategy specified",
		))
		return
	}

	_, _ = c.Writer.WriteString(marshalResponse(
		http.StatusOK,
		"",
		struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		}{fmt.Sprint(botId), name},
	))

	botId++
	mu.Unlock()
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
