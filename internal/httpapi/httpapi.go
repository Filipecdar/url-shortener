package httpapi

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	mutex            sync.RWMutex
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

	router.Route("/api", func(r chi.Router) {
		r.Post("/links", service.HandleCreateLink)
	})

	router.Get("/{shortcode}", service.HandleRedirect)

	return router
}

func (service *Service) HandleCreateLink(writer http.ResponseWriter, request *http.Request) {
	var createLinkRequest CreateLinkRequest
	if err := json.NewDecoder(request.Body).Decode(&createLinkRequest); err != nil {
		http.Error(writer, "invalid json body", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(createLinkRequest.URL)
	if validationErr := ValidateURL(longURL); validationErr != nil {
		http.Error(writer, validationErr.Error(), http.StatusBadRequest)
		return
	}

	service.mutex.Lock()
	shortcode := fmt.Sprintf("id%d", len(service.shortenedURLData)+1)
	service.shortenedURLData[shortcode] = longURL
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

func (service *Service) HandleRedirect(writer http.ResponseWriter, request *http.Request) {
	shortcode := chi.URLParam(request, "shortcode")
	if shortcode == "" {
		http.NotFound(writer, request)
		return
	}

	service.mutex.RLock()
	originalURL, exists := service.shortenedURLData[shortcode]
	service.mutex.RUnlock()

	if !exists {
		http.NotFound(writer, request)
		return
	}

	http.Redirect(writer, request, originalURL, http.StatusFound)
}

func ValidateURL(input string) error {
	if input == "" {
		return fmt.Errorf("url is required")
	}

	parsed, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("url is not a valid URI")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}

	if parsed.Host == "" {
		return fmt.Errorf("url must contain a host")
	}

	if strings.HasPrefix(parsed.Host, ":") {
		return fmt.Errorf("url host is invalid")
	}

	host := parsed.Host
	if h, _, splitErr := net.SplitHostPort(parsed.Host); splitErr == nil {
		host = h
	}
	if ip := net.ParseIP(host); ip == nil {
	}

	return nil
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
