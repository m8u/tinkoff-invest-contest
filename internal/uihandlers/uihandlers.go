package uihandlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tinkoff-invest-contest/internal/app"
)

func BotControls(c *gin.Context) {
	id := c.Query("id")
	app.Bots.Lock.RLock()
	bot, ok := app.Bots.Table[id]
	app.Bots.Lock.RUnlock()
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

func BotDescription(c *gin.Context) {
	id := c.Query("id")
	desc := app.Bots.Table[id].GetYAML()
	templateArgs := struct {
		DescriptionYAML string
	}{desc}
	c.HTML(http.StatusOK, "bot_description.html", templateArgs)
}

func CombatAccounts(c *gin.Context) {
	c.HTML(http.StatusOK, "combat_accounts.html", nil)
}

func SandboxAccounts(c *gin.Context) {
	c.HTML(http.StatusOK, "sandbox_accounts.html", nil)
}
