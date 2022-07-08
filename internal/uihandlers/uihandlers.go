package uihandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
)

func BotControls(c *gin.Context) {
	id := c.Query("id")
	bot := app.Bots[id]

	templateArgs := struct {
		Id        string
		IsStarted bool
		IsPaused  bool
	}{
		Id:        id,
		IsStarted: bot.IsStarted(),
		IsPaused:  bot.IsPaused(),
	}

	c.HTML(http.StatusOK, "bot_controls.html", templateArgs)
}
