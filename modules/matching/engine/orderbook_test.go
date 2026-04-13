package engine

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrder(id, userID string, side Side, price, qty float64) *BookOrder {
	return &BookOrder{
		ID:        id,
		UserID:    userID,
		PairID:    "BTC_USDT",
		Side:      side,
		Price:     decimal.NewFromFloat(price),
		Quantity:  decimal.NewFromFloat(qty),
		Remaining: decimal.NewFromFloat(qty),
		CreatedAt: time.Now(),
	}
}

func TestOrderBook_Add_BestBidAsk(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")

	ob.Add(newOrder("b1", "u1", SideBuy, 50000, 1))
	ob.Add(newOrder("b2", "u2", SideBuy, 50100, 1))
	ob.Add(newOrder("a1", "u3", SideSell, 50200, 1))
	ob.Add(newOrder("a2", "u4", SideSell, 50300, 1))

	assert.True(t, ob.BestBid().Equal(decimal.NewFromFloat(50100)))
	assert.True(t, ob.BestAsk().Equal(decimal.NewFromFloat(50200)))
	assert.Equal(t, 4, ob.OrderCount())
}

func TestOrderBook_EmptyBestBidAsk(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	assert.Nil(t, ob.BestBid())
	assert.Nil(t, ob.BestAsk())
}

func TestOrderBook_MatchSimple(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ask := newOrder("a1", "seller", SideSell, 50000, 0.5)
	ob.Add(ask)

	buy := newOrder("b1", "buyer", SideBuy, 50000, 0.5)
	trades := ob.Match(buy)

	require.Len(t, trades, 1)
	assert.True(t, trades[0].Quantity.Equal(decimal.NewFromFloat(0.5)))
	assert.True(t, trades[0].Price.Equal(decimal.NewFromFloat(50000)))
	assert.Equal(t, "b1", trades[0].BuyOrderID)
	assert.Equal(t, "a1", trades[0].SellOrderID)
	assert.True(t, buy.Remaining.IsZero())
	assert.Equal(t, 0, ob.OrderCount())
}

func TestOrderBook_MatchPartialFill(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "seller", SideSell, 50000, 0.5))

	buy := newOrder("b1", "buyer", SideBuy, 50000, 1.0)
	trades := ob.Match(buy)

	require.Len(t, trades, 1)
	assert.True(t, trades[0].Quantity.Equal(decimal.NewFromFloat(0.5)))
	assert.True(t, buy.Remaining.Equal(decimal.NewFromFloat(0.5)))
	assert.Equal(t, 0, ob.OrderCount()) // ask fully filled, buy not yet added
}

func TestOrderBook_MatchMultipleLevels(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "s1", SideSell, 50000, 0.3))
	ob.Add(newOrder("a2", "s2", SideSell, 50100, 0.2))
	ob.Add(newOrder("a3", "s3", SideSell, 50200, 0.5))

	buy := newOrder("b1", "buyer", SideBuy, 50100, 0.4)
	trades := ob.Match(buy)

	require.Len(t, trades, 2)
	// First fills 0.3 @ 50000, then 0.1 @ 50100
	assert.True(t, trades[0].Quantity.Equal(decimal.NewFromFloat(0.3)))
	assert.True(t, trades[0].Price.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, trades[1].Quantity.Equal(decimal.NewFromFloat(0.1)))
	assert.True(t, trades[1].Price.Equal(decimal.NewFromFloat(50100)))
	assert.True(t, buy.Remaining.IsZero())
	// a3 at 50200 untouched, a2 partial 0.1 remaining
	assert.Equal(t, 2, ob.OrderCount())
}

func TestOrderBook_NoMatch_PriceNotCross(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "seller", SideSell, 50200, 1))

	buy := newOrder("b1", "buyer", SideBuy, 50000, 1)
	trades := ob.Match(buy)

	assert.Empty(t, trades)
	assert.True(t, buy.Remaining.Equal(decimal.NewFromFloat(1)))
}

func TestOrderBook_SellMatch(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("b1", "buyer", SideBuy, 50000, 0.5))

	sell := newOrder("a1", "seller", SideSell, 50000, 0.5)
	trades := ob.Match(sell)

	require.Len(t, trades, 1)
	assert.Equal(t, "b1", trades[0].BuyOrderID)
	assert.Equal(t, "a1", trades[0].SellOrderID)
}

func TestOrderBook_SelfTradePrevention(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "same_user", SideSell, 50000, 1))

	buy := newOrder("b1", "same_user", SideBuy, 50000, 1)
	trades := ob.Match(buy)

	assert.Empty(t, trades)
	assert.True(t, buy.Remaining.Equal(decimal.NewFromFloat(1)))
}

func TestOrderBook_SelfTradePrevention_SkipToNext(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "same_user", SideSell, 50000, 0.5))
	ob.Add(newOrder("a2", "other_user", SideSell, 50000, 0.5))

	buy := newOrder("b1", "same_user", SideBuy, 50000, 0.5)
	trades := ob.Match(buy)

	require.Len(t, trades, 1)
	assert.Equal(t, "a2", trades[0].SellOrderID) // skipped a1 (same user)
}

func TestOrderBook_Cancel(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("b1", "u1", SideBuy, 50000, 1))
	assert.Equal(t, 1, ob.OrderCount())

	err := ob.Cancel("b1")
	assert.NoError(t, err)
	assert.Equal(t, 0, ob.OrderCount())
	assert.Nil(t, ob.BestBid())
}

func TestOrderBook_CancelNotFound(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	err := ob.Cancel("nonexistent")
	assert.Equal(t, ErrOrderNotFound, err)
}

func TestOrderBook_FIFO_SamePrice(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "s1", SideSell, 50000, 0.5))
	ob.Add(newOrder("a2", "s2", SideSell, 50000, 0.5))

	buy := newOrder("b1", "buyer", SideBuy, 50000, 0.5)
	trades := ob.Match(buy)

	require.Len(t, trades, 1)
	assert.Equal(t, "a1", trades[0].SellOrderID) // a1 was first (FIFO)
}

func TestOrderBook_BidDescending(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("b1", "u1", SideBuy, 49000, 1))
	ob.Add(newOrder("b2", "u2", SideBuy, 51000, 1))
	ob.Add(newOrder("b3", "u3", SideBuy, 50000, 1))

	// Best bid should be highest
	assert.True(t, ob.BestBid().Equal(decimal.NewFromFloat(51000)))

	// Sell should match against highest bid first
	sell := newOrder("a1", "seller", SideSell, 49000, 1.5)
	trades := ob.Match(sell)

	require.Len(t, trades, 2)
	assert.True(t, trades[0].Price.Equal(decimal.NewFromFloat(51000))) // highest first
	assert.True(t, trades[1].Price.Equal(decimal.NewFromFloat(50000)))
}

func TestOrderBook_AskAscending(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "s1", SideSell, 51000, 1))
	ob.Add(newOrder("a2", "s2", SideSell, 49000, 1))
	ob.Add(newOrder("a3", "s3", SideSell, 50000, 1))

	// Best ask should be lowest
	assert.True(t, ob.BestAsk().Equal(decimal.NewFromFloat(49000)))
}

func TestOrderBook_Snapshot_Empty(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	snap := ob.Snapshot()

	assert.Equal(t, "BTC_USDT", snap.PairID)
	assert.Empty(t, snap.Bids)
	assert.Empty(t, snap.Asks)
}

func TestOrderBook_Snapshot_AggregatesLevels(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")

	ob.Add(newOrder("b1", "u1", SideBuy, 50000, 1.0))
	ob.Add(newOrder("b2", "u2", SideBuy, 50000, 0.5))
	ob.Add(newOrder("b3", "u3", SideBuy, 49900, 2.0))
	ob.Add(newOrder("a1", "u4", SideSell, 50100, 0.3))

	snap := ob.Snapshot()

	require.Len(t, snap.Bids, 2)
	assert.True(t, snap.Bids[0].Price.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, snap.Bids[0].Quantity.Equal(decimal.NewFromFloat(1.5)))
	assert.Equal(t, 2, snap.Bids[0].OrderCount)
	assert.True(t, snap.Bids[1].Price.Equal(decimal.NewFromFloat(49900)))
	assert.Equal(t, 1, snap.Bids[1].OrderCount)

	require.Len(t, snap.Asks, 1)
	assert.True(t, snap.Asks[0].Price.Equal(decimal.NewFromFloat(50100)))
	assert.Equal(t, 1, snap.Asks[0].OrderCount)
}

func TestOrderBook_Snapshot_ReflectsPartialFill(t *testing.T) {
	ob := NewOrderBook("BTC_USDT")
	ob.Add(newOrder("a1", "u1", SideSell, 50000, 1.0))

	incoming := newOrder("b1", "u2", SideBuy, 50000, 0.3)
	trades := ob.Match(incoming)
	require.Len(t, trades, 1)

	snap := ob.Snapshot()
	require.Len(t, snap.Asks, 1)
	assert.True(t, snap.Asks[0].Quantity.Equal(decimal.NewFromFloat(0.7)))
	assert.Empty(t, snap.Bids)
}
