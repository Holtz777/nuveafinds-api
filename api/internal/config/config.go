// Package config carrega configuração a partir de variáveis de ambiente.
// Mantemos simples: sem biblioteca externa, só os.Getenv + defaults.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config agrupa todas as configs da API.
type Config struct {
	Port string

	OpenRouterAPIKey  string
	OpenRouterModel   string
	OpenRouterReferer string
	OpenRouterTitle   string

	PinterestAccessToken   string
	PinterestSandboxToken string
	PinterestClientID     string
	PinterestClientSecret string
	PinterestRefreshToken string
	PinterestSandbox      bool
	// BoardMap: chave = slug interno (ex: "viral-makeup-skincare-finds"),
	// valor = board ID real do Pinterest.
	BoardMap map[string]string
	// BoardMapInverse: chave = board ID, valor = slug (gerado automaticamente).
	BoardMapInverse map[string]string

	CORSOrigin string
}

// Load lê as variáveis de ambiente e devolve a Config preenchida.
// Falha se as obrigatórias estiverem vazias (no boot do server).
func Load() (*Config, error) {
	loadDotEnv()

	c := &Config{
		Port:                  getenv("PORT", "8080"),
		OpenRouterAPIKey:      os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel:       getenv("OPENROUTER_MODEL", "anthropic/claude-sonnet-4"),
		OpenRouterReferer:     getenv("OPENROUTER_REFERER", "https://nuveafinds.com"),
		OpenRouterTitle:       getenv("OPENROUTER_TITLE", "Nuvea Finds API"),
		PinterestAccessToken:   os.Getenv("PINTEREST_ACCESS_TOKEN"),
		PinterestSandboxToken: os.Getenv("PINTEREST_SANDBOX_TOKEN"),
		PinterestClientID:     os.Getenv("PINTEREST_CLIENT_ID"),
		PinterestClientSecret: os.Getenv("PINTEREST_CLIENT_SECRET"),
		PinterestRefreshToken: os.Getenv("PINTEREST_REFRESH_TOKEN"),
		PinterestSandbox:      os.Getenv("PINTEREST_SANDBOX") == "true",
		BoardMap:              parseBoardMap(os.Getenv("PINTEREST_BOARD_MAP")),
		CORSOrigin:            getenv("CORS_ORIGIN", "*"),
	}

	if c.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	if c.PinterestAccessToken == "" {
		fmt.Fprintln(os.Stderr, "WARNING: PINTEREST_ACCESS_TOKEN not set — /pin-upload (AI) works, but Pinterest endpoints will fail")
	}

	c.BoardMapInverse = make(map[string]string, len(c.BoardMap))
	for slug, id := range c.BoardMap {
		c.BoardMapInverse[id] = slug
	}

	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

// parseBoardMap converte "slug1=id1,slug2=id2" em map[string]string.
func parseBoardMap(raw string) map[string]string {
	m := make(map[string]string)
	if raw == "" {
		return m
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return m
}
