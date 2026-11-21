package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spec-kit-relm/inventory-service/internal/domain"
)

func TestAvailability(t *testing.T) {
	srv := httptest.NewServer(NewServer(domain.NewStock(map[string]int{"SKU-1": 5}), slog.New(slog.NewTextHandler(io.Discard, nil))))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/v1/items/SKU-1/availability")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !strings.Contains(string(body), `"available":5`) {
		t.Fatalf("got %d %s", res.StatusCode, body)
	}
	if res.Header.Get("X-Request-Id") == "" {
		t.Fatal("missing X-Request-Id (BE-1)")
	}

	res, _ = http.Get(srv.URL + "/v1/items/NOPE/availability")
	if res.StatusCode != 404 || res.Header.Get("Content-Type") != "application/problem+json" {
		t.Fatalf("want 404 problem+json, got %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
}
