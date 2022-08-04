package utils

import (
	"errors"
	"log"
	"net/http"
	"time"
	"tinkoff-invest-contest/internal/appstate"
)

// WaitForInternetConnection pings clients3.google.com, blocking current goroutine until connection successful
func WaitForInternetConnection() {
	httpClient := http.Client{Timeout: 5 * time.Second}
	err := errors.New("")
	for err != nil {
		_, err = httpClient.Get("https://clients3.google.com/")
		if err != nil {
			if !appstate.NoInternetConnection {
				log.Println("waiting for internet connection...")
			}
			appstate.NoInternetConnection = true
			time.Sleep(10 * time.Second)
		}
	}
	if appstate.NoInternetConnection {
		log.Println("internet connection established")
	}
	appstate.NoInternetConnection = false
}
