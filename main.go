package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Filipecdar/url-shortener/internal/config"
	"github.com/Filipecdar/url-shortener/internal/httpapi"
	"github.com/Filipecdar/url-shortener/internal/store"
)

func main() {
	applicationConfig := config.FromEnv()

	// Criar store do Postgres.
	rootContext := context.Background()
	urlStore, err := store.NewPostgresURLStore(rootContext, applicationConfig.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer urlStore.Close()

	// Criar serviço HTTP com baseURL e store.
	urlShortenerService := httpapi.NewService(applicationConfig.PublicBaseURL, urlStore)
	httpRouter := urlShortenerService.Routes()

	serverAddress := ":" + applicationConfig.Port
	log.Println("Server starting on", serverAddress, "public:", applicationConfig.PublicBaseURL)

	if err := http.ListenAndServe(serverAddress, httpRouter); err != nil {
		log.Fatal(err)
	}
}
