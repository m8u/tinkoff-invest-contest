package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"tinkoff-invest-contest/internal/strategies"
)

func GetStrategiesNames(c *gin.Context) {
	s, _ := json.Marshal(strategies.Names)
	_, _ = c.Writer.WriteString(string(s))
}

func GetStrategyDefaults(c *gin.Context) {
	name := c.Query("name")
	var s string
	if defaults, ok := strategies.DefaultsJSON[name]; ok {
		s = defaults()
	}
	_, _ = c.Writer.WriteString(s)
}
