package utils

import (
	"log"
	"os"
)

func GetSandboxToken() string {
	token := os.Getenv("SANDBOX_TOKEN")
	if token == "" {
		log.Fatalln("please provide sandbox token via 'SANDBOX_TOKEN' environment variable")
	}
	return token
}

func GetCombatToken() string {
	token := os.Getenv("COMBAT_TOKEN")
	if token == "" {
		log.Fatalln("please provide combat token via 'COMBAT_TOKEN' environment variable")
	}
	return token
}

func GetGrafanaToken() string {
	token := os.Getenv("GRAFANA_TOKEN")
	if token == "" {
		log.Println("please provide Grafana admin token via 'GRAFANA_TOKEN' environment variable")
	}
	return token
}
