package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"tinkoff-invest-contest/internal/app"
	"tinkoff-invest-contest/internal/bots"
	"tinkoff-invest-contest/internal/strategies/tistrategy"
	"tinkoff-invest-contest/internal/tradeenv"
	"tinkoff-invest-contest/internal/utils"
)

var botId int64

func CreateBot(c *gin.Context) {
	args := struct {
		Sandbox        bool   `form:"sandbox"`
		Figi           string `form:"figi"`
		InstrumentType string `form:"instrumentType"`
		AllowMargin    bool   `form:"allowMargin"`

		Strategy       string `form:"strategy"` // example: &strategy={"name":"bollinger","coef":3,"pointDev":0.001}
		CandleInterval string `form:"candleInterval"`
		Window         int    `form:"window"`
	}{}

	err := c.Bind(&args)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}

	var tradeEnv *tradeenv.TradeEnv
	if args.Sandbox {
		if app.SandboxEnv == nil {
			app.SandboxEnv = tradeenv.New(utils.GetSandboxToken(), true, utils.Fees[utils.Trader])
		}
		tradeEnv = app.SandboxEnv
	} else {
		if app.CombatEnv == nil {
			app.CombatEnv = tradeenv.New(utils.GetCombatToken(), false, 0)
		}
		tradeEnv = app.CombatEnv
	}
	instrumentType, err := utils.InstrumentTypeFromString(args.InstrumentType)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}

	// Starting from here, determine strategy type (candles/orderbook) from its name,
	// then call a function which creates a corresponding bot.
	strategyParams := struct {
		Name string `json:"name"`
	}{}
	err = json.Unmarshal([]byte(args.Strategy), &strategyParams)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	if newStrategyFromJson, ok := tistrategy.JsonConstructors[strategyParams.Name]; ok {
		strategy, err := newStrategyFromJson(args.Strategy)
		if err != nil {
			c.Writer.WriteString(err.Error())
			return
		}
		candleInterval, err := utils.CandleIntervalFromString(args.CandleInterval)
		if err != nil {
			c.Writer.WriteString(err.Error())
			return
		}
		id := fmt.Sprint(botId)
		botId++
		app.Bots[id] = bots.NewTechnicalIndicatorBot(tradeEnv, args.Figi, instrumentType, candleInterval, args.Window, args.AllowMargin, strategy)
	} // else if newStrategyFromJson, ok := obstrategy.JsonConstructors[strategyParams.Name]; ok {

	//}

	c.Writer.WriteString("ok")
}
