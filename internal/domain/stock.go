// Package domain holds inventory business rules. It has no imports from transport or storage
// packages (INV-2): adapters depend on the domain, never the other way round.
package domain

import (
	"errors"
	"sync"
)

// ErrUnknownSKU is returned when a SKU has never been stocked.
var ErrUnknownSKU = errors.New("unknown sku")

// Stock tracks on-hand units per SKU. Available units can never be negative (INV-3).
type Stock struct {
	mu     sync.RWMutex
	onHand map[string]int
}

// NewStock returns a stock ledger seeded with on-hand quantities.
func NewStock(seed map[string]int) *Stock {
	s := &Stock{onHand: map[string]int{}}
	for sku, qty := range seed {
		if qty < 0 {
			qty = 0
		}
		s.onHand[sku] = qty
	}
	return s
}

// Available returns units that can be sold for a SKU.
func (s *Stock) Available(sku string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	qty, ok := s.onHand[sku]
	if !ok {
		return 0, ErrUnknownSKU
	}
	return qty, nil
}
