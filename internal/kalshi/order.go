package kalshi

import "context"

type Side string
type Action string

const (
	SideYes Side = "yes"
	SideNo  Side = "no"

	ActionBuy  Action = "buy"
	ActionSell Action = "sell"
)

// OrderIntent is a proposed Kalshi order produced by a strategy, before
// risk checks and execution
type OrderIntent struct {
	Ticker     string
	Side       Side
	Action     Action
	Count      int
	LimitPrice float64
}

type OrderStatus string

const (
	OrderResting  OrderStatus = "resting"
	OrderFilled   OrderStatus = "filled"
	OrderCanceled OrderStatus = "canceled"
)

// OrderResult is what a broker returns once an order is accepted (ack),
// before it fills. VenueOrderID is the id the venue assigns.
type OrderResult struct {
	VenueOrderID string
	Status       OrderStatus
}

// Broker is the Kalshi order contract. orderID is the client-generated id
// (also sent as the venue's idempotency token); the result carries the
// venue's own id once it acks.
type Broker interface {
	PlaceOrder(ctx context.Context, intent OrderIntent, orderID string) (OrderResult, error)
}
