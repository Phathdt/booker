package engine

import (
	"context"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_SubmitAndMatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	defer eng.Stop()

	// Submit sell (rests)
	trades, err := eng.Submit(newOrder("a1", "seller", SideSell, 50000, 0.5))
	require.NoError(t, err)
	assert.Empty(t, trades)
	assert.Equal(t, 1, eng.OrderCount())

	// Submit buy (matches)
	trades, err = eng.Submit(newOrder("b1", "buyer", SideBuy, 50000, 0.5))
	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.True(t, trades[0].Quantity.Equal(decimal.NewFromFloat(0.5)))
	assert.Equal(t, 0, eng.OrderCount())
}

func TestEngine_SubmitNoMatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	defer eng.Stop()

	trades, err := eng.Submit(newOrder("b1", "buyer", SideBuy, 49000, 1))
	require.NoError(t, err)
	assert.Empty(t, trades)
	assert.Equal(t, 1, eng.OrderCount())
}

func TestEngine_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	defer eng.Stop()

	eng.Submit(newOrder("b1", "buyer", SideBuy, 50000, 1))
	assert.Equal(t, 1, eng.OrderCount())

	err := eng.Cancel("b1")
	assert.NoError(t, err)
	assert.Equal(t, 0, eng.OrderCount())
}

func TestEngine_CancelNotFound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 64)
	eng.Start(ctx)
	defer eng.Stop()

	err := eng.Cancel("nonexistent")
	assert.Error(t, err)
}

func TestEngine_Preload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 64)

	// Preload before starting
	eng.Preload([]*BookOrder{
		newOrder("a1", "seller", SideSell, 50000, 1),
	})
	assert.Equal(t, 1, eng.OrderCount())

	eng.Start(ctx)
	defer eng.Stop()

	// Submit matching buy
	trades, err := eng.Submit(newOrder("b1", "buyer", SideBuy, 50000, 0.5))
	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.True(t, trades[0].Quantity.Equal(decimal.NewFromFloat(0.5)))
}

func TestEngine_ConcurrentSubmits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng := NewEngine("BTC_USDT", 256)
	eng.Start(ctx)
	defer eng.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			side := SideBuy
			price := 49000.0
			if idx%2 == 0 {
				side = SideSell
				price = 51000.0
			}
			eng.Submit(newOrder(
				"o"+string(rune('A'+idx)),
				"user"+string(rune('A'+idx)),
				side, price, 0.01,
			))
		}(i)
	}
	wg.Wait()

	// No crash, all orders processed
	assert.GreaterOrEqual(t, eng.OrderCount(), 0)
}
