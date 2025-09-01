package config

import "os"

type Config struct {
	Port          string
	PublicBaseURL string
}

func FromEnv() Config {
	cfg := Config{
		Port:          getenv("PORT", "8080"),
		PublicBaseURL: getenv("PUBLIC_BASE_URL", "http://localhost:8080"),
	}
	return cfg
}

func getenv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
