package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Filipecdar/url-shortener/internal/short"
	"github.com/Filipecdar/url-shortener/internal/store"
)

// CreateLinkRequest representa o payload esperado para criação de link.
type CreateLinkRequest struct {
	URL string `json:"url"`
}

// CreateLinkResponse representa a resposta da criação de link.
type CreateLinkResponse struct {
	Shortcode string `json:"shortcode"`
	ShortURL  string `json:"shortUrl"`
}

// Service é o serviço principal de encurtamento de URLs.
type Service struct {
	baseURL  string
	urlStore store.URLStore
}

// NewService cria e inicializa um novo Service.
func NewService(baseURL string, urlStore store.URLStore) *Service {
	return &Service{
		baseURL:  baseURL,
		urlStore: urlStore,
	}
}

// Routes retorna o roteador HTTP com as rotas registradas.
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

// HandleCreateLink processa a criação de um novo link curto.
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

	ctx := request.Context()
	newID, err := service.urlStore.InsertURL(ctx, longURL)
	if err != nil {
		http.Error(writer, "failed to persist url", http.StatusInternalServerError)
		return
	}

	shortcode := short.EncodeBase62(newID)

	response := CreateLinkResponse{
		Shortcode: shortcode,
		ShortURL:  fmt.Sprintf("%s/%s", removeTrailingSlash(service.baseURL), shortcode),
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(response); err != nil {
		http.Error(writer, "failed to encode response", http.StatusInternalServerError)
	}
}

// HandleRedirect resolve o shortcode e redireciona para a URL original.
func (service *Service) HandleRedirect(writer http.ResponseWriter, request *http.Request) {
	shortcode := chi.URLParam(request, "shortcode")
	if shortcode == "" {
		http.NotFound(writer, request)
		return
	}

	id, err := short.DecodeBase62(shortcode)
	if err != nil {
		http.NotFound(writer, request)
		return
	}

	ctx := request.Context()
	record, err := service.urlStore.GetURLByID(ctx, id)
	if err != nil || record == nil {
		http.NotFound(writer, request)
		return
	}

	http.Redirect(writer, request, record.LongURL, http.StatusFound)
}

// ValidateURL aplica validações básicas em uma URL de entrada.
func ValidateURL(input string) error {
	if input == "" {
		return fmt.Errorf("url is required")
	}

	parsed, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("url is not a valid URI")
	}

	// Exigir esquema http(s).
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}

	// Exigir host (domínio ou IP).
	if parsed.Host == "" {
		return fmt.Errorf("url must contain a host")
	}

	// Rejeitar formatos como ":8080" (porta sem host).
	if strings.HasPrefix(parsed.Host, ":") {
		return fmt.Errorf("url host is invalid")
	}

	// Separar host e porta, quando houver, para validar somente o host puro.
	host := parsed.Host
	if h, _, splitErr := net.SplitHostPort(parsed.Host); splitErr == nil {
		host = h
	}

	// Se for IP, validar; se não for IP, assumimos que é domínio (ok).
	if ip := net.ParseIP(host); ip == nil {
		// não é IP → provavelmente domínio; ok.
		// (Aqui poderíamos validar TLDs, bloquear localhost, etc., se quisermos apertar.)
	}

	return nil
}

// removeTrailingSlash remove a barra final de uma URL base, se existir.
func removeTrailingSlash(input string) string {
	if input == "" {
		return input
	}
	if input[len(input)-1] == '/' {
		return input[:len(input)-1]
	}
	return input
}

// (Opcional) contexto com timeout padrão para chamadas ao store.
func contextWithTimeout(parent context.Context) (context.Context, func()) {
	// Mantido aqui para eventual uso futuro (timeouts por rota).
	return context.WithCancel(parent)
}
