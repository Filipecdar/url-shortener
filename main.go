package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/Filipecdar/url-shortener/internal/config"
)

func main() {
	cfg := config.FromEnv()

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	addr := ":" + cfg.Port
	log.Println("listening on", addr, "public:", cfg.PublicBaseURL)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
