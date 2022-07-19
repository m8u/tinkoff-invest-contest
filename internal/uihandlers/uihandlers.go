package uihandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
)

func BotControls(c *gin.Context) {
	id := c.Query("id")
	app.Bots.Lock.Lock()
	bot, ok := app.Bots.Table[id]
	app.Bots.Lock.Unlock()
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

func BotLogConsole(c *gin.Context) {
	id := c.Query("id")

	templateArgs := struct {
		Id string
	}{id}

	c.HTML(http.StatusOK, "bot_log.html", templateArgs)
}
