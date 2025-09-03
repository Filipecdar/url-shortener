package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

type CreateLinkRequest struct {
	URL string `json:"url"`
}

type CreateLinkResponse struct {
	Shortcode string `json:"shortcode"`
	ShortURL  string `json:"shortUrl"`
}

type Service struct {
	baseURL          string
	mutex            sync.Mutex
	shortenedURLData map[string]string
}

func NewService(baseURL string) *Service {
	return &Service{
		baseURL:          baseURL,
		shortenedURLData: make(map[string]string),
	}
}

func (service *Service) Routes() http.Handler {
	router := chi.NewRouter()

	router.Get("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		fmt.Fprint(writer, "ok")
	})

	router.Post("/api/links", service.HandleCreateLink)

	return router
}

func (service *Service) HandleCreateLink(writer http.ResponseWriter, request *http.Request) {
	var createLinkRequest CreateLinkRequest
	if err := json.NewDecoder(request.Body).Decode(&createLinkRequest); err != nil || createLinkRequest.URL == "" {
		http.Error(writer, "invalid json", http.StatusBadRequest)
		return
	}

	service.mutex.Lock()
	shortcode := fmt.Sprintf("id%d", len(service.shortenedURLData)+1)
	service.shortenedURLData[shortcode] = createLinkRequest.URL
	service.mutex.Unlock()

	createLinkResponse := CreateLinkResponse{
		Shortcode: shortcode,
		ShortURL:  fmt.Sprintf("%s/%s", removeTrailingSlash(service.baseURL), shortcode),
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(createLinkResponse); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
	}
}

func removeTrailingSlash(input string) string {
	if input == "" {
		return input
	}
	if input[len(input)-1] == '/' {
		return input[:len(input)-1]
	}
	return input
}
