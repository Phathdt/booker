package trades

import "sync"

// TradeInfo represents a single trade for display.
type TradeInfo struct {
	TradeID   string `json:"trade_id"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	BuyerID   string `json:"-"`
	SellerID  string `json:"-"`
	Timestamp int64  `json:"timestamp"` // unix ms
}

const maxTrades = 100

// RecentTrades maintains a circular buffer of recent trades for a pair.
type RecentTrades struct {
	mu     sync.RWMutex
	trades []TradeInfo
}

func NewRecentTrades() *RecentTrades {
	return &RecentTrades{
		trades: make([]TradeInfo, 0, maxTrades),
	}
}

// Add appends a trade, evicting the oldest if at capacity.
func (rt *RecentTrades) Add(trade TradeInfo) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.trades) >= maxTrades {
		rt.trades = rt.trades[1:]
	}
	rt.trades = append(rt.trades, trade)
}

// GetRecent returns the last N trades (newest first).
func (rt *RecentTrades) GetRecent(limit int) []TradeInfo {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if limit <= 0 || limit > len(rt.trades) {
		limit = len(rt.trades)
	}

	// Return newest first
	result := make([]TradeInfo, limit)
	for i := 0; i < limit; i++ {
		result[i] = rt.trades[len(rt.trades)-1-i]
	}
	return result
}
