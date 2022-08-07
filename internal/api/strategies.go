package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"tinkoff-invest-contest/internal/strategies/strategy"
)

func GetStrategiesNames(c *gin.Context) {
	s, _ := json.Marshal(strategy.Names)
	_, _ = c.Writer.WriteString(string(s))
}

func GetStrategyDefaults(c *gin.Context) {
	name := c.Query("name")
	var s string
	if defaults, ok := strategy.DefaultsJSON[name]; ok {
		s = defaults()
	}
	_, _ = c.Writer.WriteString(s)
}
