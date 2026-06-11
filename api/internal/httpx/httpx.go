// Package httpx contém helpers para resposta JSON, logging e middleware.
// Evita dependências externas - só net/http + encoding/json.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// JSON escreve um payload como JSON com o status HTTP dado.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("httpx: encode error: %v", err)
	}
}

// Error responde um JSON padronizado de erro.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{
		"status":  "error",
		"message": message,
	})
}

// Success responde um JSON padronizado de sucesso com os dados.
func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, map[string]any{
		"status": "success",
		"data":   data,
	})
}

// Logger é um middleware que registra método, path, status e duração.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

// CORS é um middleware simples. Aceita um origin ou "*".
func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter captura o status code para logging.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
