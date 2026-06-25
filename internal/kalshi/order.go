package kalshi

type Side string
type Action string

const (
	SideYes Side = "yes"
	SideNo  Side = "no"

	ActionBuy  Action = "buy"
	ActionSell Action = "sell"
)

// OrderIntent is a proposed Kalshi order produced by a strategy, before
// risk checks and execution. LimitPrice is in dollars (0–1).
type OrderIntent struct {
	Ticker     string
	Side       Side
	Action     Action
	Count      int
	LimitPrice float64
}
