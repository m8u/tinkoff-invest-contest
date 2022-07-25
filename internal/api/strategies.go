package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"tinkoff-invest-contest/internal/strategies"
	"tinkoff-invest-contest/internal/strategies/tistrategy"
)

func GetStrategiesNames(c *gin.Context) {
	s, _ := json.Marshal(strategies.Names)
	_, _ = c.Writer.WriteString(string(s))
}

func GetStrategyDefaults(c *gin.Context) {
	name := c.Query("name")
	var s string
	if defaults, ok := tistrategy.DefaultsJSON[name]; ok {
		s = defaults()
	} // else if defaults, ok := obstrategy.DefaultsJSON[name]; ok {
	_, _ = c.Writer.WriteString(s)
}
