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
var botDashboardTemplate []byte

func InitGrafana() {
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
	}
	for _, dataSource := range dataSources {
		if dataSource.Name == "PostgreSQL" {
			err := client.DeleteDataSource(dataSource.ID)
			if err != nil {
				log.Printf("can't delete Grafana datasource: %v", err)
			}
		}
	}

	_, err = client.NewDataSource(&grafana.DataSource{
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
	if err != nil {
		log.Printf("can't add Grafana datasource: %v", err)
		client = nil
	}

	botDashboardTemplate, err = os.ReadFile("internal/dashboard/templates/bot_dashboard.json")
	if err != nil {
		log.Fatalf("can't read template file: %v", err)
	}
}

func IsGrafanaInitialized() bool {
	if client == nil {
		log.Printf("Grafana API client was not initialized")
		return false
	}
	return true
}

func AddBotDashboard(figi string) error {
	if !IsGrafanaInitialized() {
		return nil
	}
	modelStr := string(botDashboardTemplate)
	modelStr = strings.ReplaceAll(modelStr, "<figi>", strings.ToLower(figi))
	var model map[string]any
	_ = json.Unmarshal([]byte(modelStr), &model)
	dashboard := grafana.Dashboard{
		Model:     model,
		Overwrite: true,
	}
	_, err := client.NewDashboard(dashboard)
	if err != nil {
		return err
	}
	return nil
}
