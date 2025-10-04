# Build
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/url-shortener ./main.go

# Runtime
FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/url-shortener /usr/local/bin/url-shortener

# Variáveis padrão (podem ser sobrescritas no Compose)
ENV PORT=8080
ENV PUBLIC_BASE_URL=http://localhost:8080
ENV DATABASE_URL=postgres://urlshort:secret@postgres:5432/urlshort?sslmode=disable

EXPOSE 8080
CMD ["url-shortener"]
