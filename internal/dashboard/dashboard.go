package dashboard

import (
	"encoding/json"
	grafana "github.com/grafana/grafana-api-golang-client"
	"log"
	"os"
	"strings"
	db "tinkoff-invest-contest/internal/database"
	"tinkoff-invest-contest/internal/utils"
)

var client *grafana.Client

var botsFolder grafana.Folder
var utilitiesFolder grafana.Folder

var botDashboardTemplate []byte

func init() {
	var err error
	client, err = grafana.New("http://grafana:3000", grafana.Config{
		APIKey:     utils.GetGrafanaToken(),
		NumRetries: 1,
	})
	if err != nil {
		log.Fatalf("error creating Grafana API client: %v", err)
	}

	dataSources, err := client.DataSources()
	if err != nil {
		log.Printf("can't get Grafana datasources: %v", err)
		client = nil // TODO: try again after some time if not initialized (when trade service starts before grafana)
		return
	}
	for _, dataSource := range dataSources {
		if dataSource.Name == "PostgreSQL" {
			_ = client.DeleteDataSource(dataSource.ID)
		}
	}

	_, _ = client.NewDataSource(&grafana.DataSource{
		Name:      "PostgreSQL",
		UID:       "PostgreSQL",
		Type:      "postgres",
		URL:       db.Host + ":5432",
		Database:  db.DBname,
		User:      db.User,
		Access:    "proxy",
		IsDefault: true,
		JSONData: grafana.JSONData{
			Sslmode:      "disable",
			TLSAuth:      false,
			TimeInterval: "1s",
		},
		SecureJSONData: grafana.SecureJSONData{
			Password: db.Password,
		},
	})

	folders, _ := client.Folders()
	for _, folder := range folders {
		_ = client.DeleteFolder(folder.UID)
	}
	botsFolder, _ = client.NewFolder("Bots")
	utilitiesFolder, _ = client.NewFolder("Utilities")

	_ = addUtilityDashboard("internal/dashboard/templates/create_bot_dashboard.json", utilitiesFolder.ID)

	botDashboardTemplate, _ = os.ReadFile("internal/dashboard/templates/bot_dashboard.json")
}

func addUtilityDashboard(templatePath string, folderID int64) error {
	template, _ := os.ReadFile(templatePath)
	modelStr := string(template)
	modelStr = strings.ReplaceAll(modelStr, "<host>", os.Getenv("HOST"))
	modelStr = strings.ReplaceAll(modelStr, "<port>", os.Getenv("PORT"))
	var model map[string]any
	_ = json.Unmarshal([]byte(modelStr), &model)
	dashboard := grafana.Dashboard{
		Model:     model,
		Folder:    folderID,
		Overwrite: true,
	}
	_, err := client.NewDashboard(dashboard)
	return err
}

func IsGrafanaInitialized() bool {
	if client == nil {
		log.Printf("Grafana API client was not initialized")
		return false
	}
	return true
}

func AddBotDashboard(botId string, botName string) error {
	if !IsGrafanaInitialized() {
		return nil
	}
	modelStr := string(botDashboardTemplate)
	modelStr = strings.ReplaceAll(modelStr, "<bot_id>", strings.ToLower(botId))
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
	_, err := client.NewDashboard(dashboard)
	if err != nil {
		return err
	}
	return nil
}

func RemoveBotDashboards() {
	err := client.DeleteFolder(botsFolder.UID)
	if err != nil {
		log.Printf("can't delete Grafana folder: %v", err)
		client = nil
	}
}
