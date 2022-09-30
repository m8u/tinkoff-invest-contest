package dashboard

import (
	"encoding/json"
	"fmt"
	grafana "github.com/grafana/grafana-api-golang-client"
	"log"
	"os"
	"strconv"
	"strings"
	"tinkoff-invest-contest/internal/client/investapi"
	"tinkoff-invest-contest/internal/utils"
)

var client *grafana.Client
var botsFolder grafana.Folder
var botDashboardTemplate []byte
var botDashboards map[int]int64

func init() {
	var err error
	client, err = grafana.New("http://grafana:3000", grafana.Config{
		APIKey:     utils.GetGrafanaToken(),
		NumRetries: 1,
	})
	if err != nil {
		log.Fatalf("error creating Grafana API client: %v", err)
	}

	folders, _ := client.Folders()
	for _, folder := range folders {
		_ = client.DeleteFolder(folder.UID)
	}
	botsFolder, _ = client.NewFolder("Bots")

	addUtilityDashboard("internal/dashboard/templates/manage_bots.json")
	addUtilityDashboard("internal/dashboard/templates/manage_accounts.json")

	botDashboardTemplate, _ = os.ReadFile("internal/dashboard/templates/bot_dashboard.json")
	botDashboards = make(map[int]int64)
}

func addUtilityDashboard(templatePath string) {
	template, _ := os.ReadFile(templatePath)
	modelStr := string(template)
	modelStr = strings.ReplaceAll(modelStr, "<host>", os.Getenv("HOST"))
	modelStr = strings.ReplaceAll(modelStr, "<port>", os.Getenv("PORT"))
	modelStr = strings.ReplaceAll(modelStr, "<bots_folder_id>", strconv.FormatInt(botsFolder.ID, 10))
	var model map[string]any
	_ = json.Unmarshal([]byte(modelStr), &model)
	dashboard := grafana.Dashboard{
		Model:     model,
		Overwrite: true,
	}
	_, err := client.NewDashboard(dashboard)
	utils.MaybeCrash(err)
}

func IsGrafanaInitialized() bool {
	if client == nil {
		log.Printf("Grafana API client was not initialized")
		return false
	}
	return true
}

func AddBotDashboard(botId int, botName string) {
	if !IsGrafanaInitialized() {
		return
	}
	modelStr := string(botDashboardTemplate)
	modelStr = strings.ReplaceAll(modelStr, "<bot_id>", fmt.Sprint(botId))
	modelStr = strings.ReplaceAll(modelStr, "<bot_name>", strings.ToLower(botName))
	modelStr = strings.ReplaceAll(modelStr, "<host>", os.Getenv("HOST"))
	modelStr = strings.ReplaceAll(modelStr, "<port>", os.Getenv("PORT"))
	var model map[string]any
	_ = json.Unmarshal([]byte(modelStr), &model)
	dashboard := grafana.Dashboard{
		Model:     model,
		Folder:    botsFolder.ID,
		Overwrite: true,
	}
	resp, err := client.NewDashboard(dashboard)
	utils.MaybeCrash(err)
	botDashboards[botId] = resp.ID
}

func RemoveBotDashboards() {
	err := client.DeleteFolder(botsFolder.UID)
	if err != nil {
		log.Printf("can't delete Grafana folder: %v", err)
		client = nil
	}
}

func AnnotateOrder(botId int, direction investapi.OrderDirection, quantity int64, price float64, currency string) error {
	_, err := client.NewAnnotation(&grafana.Annotation{
		DashboardID: botDashboards[botId],
		PanelID:     0,
		Text: fmt.Sprintf("%v %v for avg. %v %v",
			utils.OrderDirectionToString(direction),
			quantity,
			price,
			currency,
		),
		Tags: []string{utils.OrderDirectionToString(direction)},
	})
	return err
}
