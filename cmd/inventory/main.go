// Command inventory runs inventory-service.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spec-kit-relm/inventory-service/internal/domain"
	"github.com/spec-kit-relm/inventory-service/internal/httpapi"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	stock := domain.NewStock(map[string]int{"SKU-RED-MUG": 5, "SKU-BLUE-MUG": 1}) // demo seed (synthetic)
	addr := envOr("ADDR", ":8081")
	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewServer(stock, log),
		ReadHeaderTimeout: 2 * time.Second, // BE-5: explicit timeouts
		WriteTimeout:      5 * time.Second,
	}
	log.Info("inventory-service listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
