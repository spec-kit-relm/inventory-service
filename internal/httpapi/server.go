// Package httpapi is the HTTP adapter for contracts/inventory-api.yaml (stdlib net/http, INV-1).
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/spec-kit-relm/inventory-service/internal/domain"
)

// Problem is an RFC 9457 problem document (BE-1).
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

const problemBase = "https://errors.relm.example/inventory/"

// NewServer wires the routes of inventory-api 1.0.0. RELM-2 adds the reservation routes.
func NewServer(stock *domain.Stock, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/items/{sku}/availability", func(w http.ResponseWriter, r *http.Request) {
		sku := r.PathValue("sku")
		qty, err := stock.Available(sku)
		if errors.Is(err, domain.ErrUnknownSKU) {
			writeProblem(w, http.StatusNotFound, "not-found", "Unknown SKU", sku)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sku": sku, "available": qty})
	})
	return withRequestID(mux, log)
}

func withRequestID(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Request-Id", id)
		log.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", id)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeProblem(w http.ResponseWriter, status int, code, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{Type: problemBase + code, Title: title, Status: status, Detail: detail})
}
