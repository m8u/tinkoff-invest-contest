package uihandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
)

func BotControls(c *gin.Context) {
	id := c.Query("id")
	bot, ok := app.Bots[id]
	if !ok {
		_, _ = c.Writer.WriteString("bot #" + id + " does not exist")
		return
	}

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

func CreateBotForm(c *gin.Context) {
	c.HTML(http.StatusOK, "create_bot.html", nil)
}

func CreateSandboxAccountForm(c *gin.Context) {
	c.HTML(http.StatusOK, "create_sandbox_account.html", nil)
}
