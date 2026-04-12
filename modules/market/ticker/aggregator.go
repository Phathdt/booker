package ticker

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const bucketsCount = 1440 // 24h × 60min

// MinuteBucket stores OHLCV data for a single minute.
type MinuteBucket struct {
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	Volume decimal.Decimal
	Minute int64 // unix minute (ts / 60)
	Active bool
}

// Ticker represents 24h aggregated market data.
type Ticker struct {
	PairID    string
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
	ChangePct decimal.Decimal
	LastPrice decimal.Decimal
	Timestamp int64 // unix ms
}

// Aggregator computes rolling 24h ticker for a single trading pair.
type Aggregator struct {
	pairID    string
	buckets   [bucketsCount]MinuteBucket
	lastPrice decimal.Decimal
	mu        sync.RWMutex
}

func NewAggregator(pairID string) *Aggregator {
	return &Aggregator{pairID: pairID}
}

// AddTrade updates the ticker with a new trade.
func (a *Aggregator) AddTrade(price, quantity decimal.Decimal, ts time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	minute := ts.Unix() / 60
	idx := int(minute % bucketsCount)

	bucket := &a.buckets[idx]
	if !bucket.Active || bucket.Minute != minute {
		// New or stale bucket — reset
		bucket.Open = price
		bucket.High = price
		bucket.Low = price
		bucket.Close = price
		bucket.Volume = quantity
		bucket.Minute = minute
		bucket.Active = true
	} else {
		// Update existing bucket
		if price.GreaterThan(bucket.High) {
			bucket.High = price
		}
		if price.LessThan(bucket.Low) {
			bucket.Low = price
		}
		bucket.Close = price
		bucket.Volume = bucket.Volume.Add(quantity)
	}

	a.lastPrice = price
}

// GetTicker returns the current 24h aggregated ticker.
func (a *Aggregator) GetTicker() *Ticker {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now().Unix() / 60
	cutoff := now - bucketsCount

	t := &Ticker{
		PairID:    a.pairID,
		LastPrice: a.lastPrice,
		Timestamp: time.Now().UnixMilli(),
	}

	// Single pass: find oldest bucket (for open), track high/low/volume
	oldestMinute := int64(0)
	for i := range a.buckets {
		b := &a.buckets[i]
		if !b.Active || b.Minute <= cutoff {
			continue
		}

		// Track oldest for open price
		if oldestMinute == 0 || b.Minute < oldestMinute {
			oldestMinute = b.Minute
			t.Open = b.Open
		}

		// High/Low
		if t.High.IsZero() || b.High.GreaterThan(t.High) {
			t.High = b.High
		}
		if t.Low.IsZero() || b.Low.LessThan(t.Low) {
			t.Low = b.Low
		}

		t.Volume = t.Volume.Add(b.Volume)
	}

	// Close = last trade price
	t.Close = a.lastPrice

	// Change percent
	if !t.Open.IsZero() {
		t.ChangePct = t.Close.Sub(t.Open).Div(t.Open).Mul(decimal.NewFromInt(100))
	}

	return t
}
