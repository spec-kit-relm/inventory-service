package domain

import "testing"

func TestAvailableNeverNegative(t *testing.T) {
	s := NewStock(map[string]int{"SKU-1": -3, "SKU-2": 5})
	if got, _ := s.Available("SKU-1"); got != 0 {
		t.Fatalf("negative seed must clamp to 0, got %d", got)
	}
	if got, _ := s.Available("SKU-2"); got != 5 {
		t.Fatalf("want 5, got %d", got)
	}
	if _, err := s.Available("NOPE"); err != ErrUnknownSKU {
		t.Fatalf("want ErrUnknownSKU, got %v", err)
	}
}
