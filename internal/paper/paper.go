package paper

import (
	"context"
	"fmt"
	"sync"

	"github.com/Krushna-B/talon/internal/kalshi"
)

// order is what the paper engine remembers about a placed order.
type order struct {
	OrderID      string
	VenueOrderID string
	Intent       kalshi.OrderIntent
	Status       kalshi.OrderStatus
}

// Broker is an in-memory fake that satisfies kalshi.Broker. Every order
// is accepted and left "resting"; nothing fills yet.
type Broker struct {
	mu     sync.Mutex
	seq    int
	orders map[string]order
}

func New() *Broker {
	return &Broker{orders: make(map[string]order)}
}

func (b *Broker) PlaceOrder(ctx context.Context, intent kalshi.OrderIntent, orderID string) (kalshi.OrderResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.seq++
	venueID := fmt.Sprintf("paper-%06d", b.seq)
	b.orders[orderID] = order{
		OrderID:      orderID,
		VenueOrderID: venueID,
		Intent:       intent,
		Status:       kalshi.OrderResting,
	}

	return kalshi.OrderResult{VenueOrderID: venueID, Status: kalshi.OrderResting}, nil
}
