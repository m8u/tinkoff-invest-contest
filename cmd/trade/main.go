package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tinkoff-invest-contest/internal/api"
	"tinkoff-invest-contest/internal/api/botlog"
	"tinkoff-invest-contest/internal/app"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/dashboard"
	"tinkoff-invest-contest/internal/uihandlers"
)

func handleExit() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	// Wait for termination signal
	<-ch
	signal.Stop(ch)
	log.Println("Exiting...")
	// Run remove sequences for bots
	app.Bots.Lock.RLock()
	for _, bot := range app.Bots.Table {
		bot.Remove()
	}
	app.Bots.Lock.RUnlock()
	// Trigger exit actions
	appstate.ShouldExit = true
	appstate.ExitActionsWG.Done()
	// Remove Grafana dashboards
	dashboard.RemoveBotDashboards()
	// Wait for all to complete before exiting
	appstate.PostExitActionsWG.Wait()
}

func runServer() {
	router := gin.Default()

	router.LoadHTMLFiles(
		"./web/templates/bot_controls.html",
		"./web/templates/create_bot.html",
		"./web/templates/create_sandbox_account.html",
		"./web/templates/bot_log.html",
		"./web/templates/bot_description.html",
		"./web/templates/combat_accounts.html",
		"./web/templates/sandbox_accounts.html",
	)

	router.POST("/api/bots/Create", api.CreateBot)
	router.POST("/api/bots/Start", api.StartBot)
	router.POST("/api/bots/TogglePause", api.TogglePauseBot)
	router.POST("/api/bots/Remove", api.RemoveBot)

	router.GET("/api/strategies/GetNames", api.GetStrategiesNames)
	router.GET("/api/strategies/GetDefaults", api.GetStrategyDefaults)

	router.POST("/api/accounts/Create", api.CreateSandboxAccount)
	router.POST("/api/accounts/Remove", api.RemoveSandboxAccount)
	router.GET("/api/accounts/GetCombatAccounts", api.GetCombatAccounts)
	router.GET("/api/accounts/GetSandboxAccounts", api.GetSandboxAccounts)

	router.GET("/ws/botlog", botlog.Echo)

	router.GET("/botcontrols", uihandlers.BotControls)
	router.GET("/createbot", uihandlers.CreateBotForm)
	router.GET("/createsandboxaccount", uihandlers.CreateSandboxAccountForm)
	router.GET("/botlog", uihandlers.BotLogConsole)
	router.GET("/botdesc", uihandlers.BotDescription)
	router.GET("/combataccounts", uihandlers.CombatAccounts)
	router.GET("/sandboxaccounts", uihandlers.SandboxAccounts)

	log.Fatalln(router.Run())
}

func main() {
	_ = godotenv.Load(".env")

	mw := io.MultiWriter(os.Stdout, botlog.Writer)
	log.SetOutput(mw)

	go runServer()

	handleExit()
}
