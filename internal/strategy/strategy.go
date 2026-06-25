package strategy

import "github.com/Krushna-B/talon/internal/kalshi"

// Strategy turns a market tick into zero or more order intents.
// Implementations may hold their own state (Option A).
type Strategy interface {
	OnTick(tick kalshi.MarketTick) []kalshi.OrderIntent
}

// CheapYes is a throwaway WIRE-TEST strategy: if YES is offered below a
// threshold, it "wants" to buy one contract. This is NOT a real edge — it
// exists only to prove the tick → strategy → intent path works end to end.
type CheapYes struct {
	MaxAsk float64
}

func (s CheapYes) OnTick(tick kalshi.MarketTick) []kalshi.OrderIntent {
	if tick.YesAsk <= 0 || tick.YesAsk >= s.MaxAsk {
		return nil // no ask, or too expensive → do nothing
	}
	return []kalshi.OrderIntent{{
		Ticker:     tick.Ticker,
		Side:       kalshi.SideYes,
		Action:     kalshi.ActionBuy,
		Count:      1,
		LimitPrice: tick.YesAsk,
	}}
}
