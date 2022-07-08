package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tinkoff-invest-contest/internal/api"
	"tinkoff-invest-contest/internal/app"
	"tinkoff-invest-contest/internal/appstate"
	"tinkoff-invest-contest/internal/dashboard"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/strategies"
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
	for _, bot := range app.Bots {
		bot.Remove()
	}
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

	router.LoadHTMLFiles("./web/templates/bot_controls.html")

	router.POST("/api/bots/Create", api.CreateBot)
	router.GET("/api/bots/Create", api.CreateBot)
	router.POST("/api/bots/Start", api.StartBot)
	router.POST("/api/bots/TogglePause", api.TogglePauseBot)
	router.POST("/api/bots/Remove", api.RemoveBot)

	router.GET("/botcontrols", uihandlers.BotControls)

	log.Fatalln(router.Run())
}

func main() {
	appstate.ExitActionsWG.Add(1)

	_ = godotenv.Load(".env")

	db.InitDB()
	dashboard.InitGrafana()
	strategies.InitConstructorsMap()
	app.Init()

	go runServer()

	handleExit()
}
