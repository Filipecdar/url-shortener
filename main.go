package main

import (
	"log"
	"net/http"

	"github.com/Filipecdar/url-shortener/internal/config"
	"github.com/Filipecdar/url-shortener/internal/httpapi"
)

func main() {
	applicationConfig := config.FromEnv()

	urlShortenerService := httpapi.NewService(applicationConfig.PublicBaseURL)
	httpRouter := urlShortenerService.Routes()

	serverAddress := ":" + applicationConfig.Port
	log.Println("Server starting on", serverAddress, "public:", applicationConfig.PublicBaseURL)

	if err := http.ListenAndServe(serverAddress, httpRouter); err != nil {
		log.Fatal(err)
	}
}
