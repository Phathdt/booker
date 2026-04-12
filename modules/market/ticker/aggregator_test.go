package ticker

import (
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAggregator_AddTrade_SingleTrade(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	now := time.Now()

	agg.AddTrade(decimal.NewFromFloat(50000), decimal.NewFromFloat(1), now)

	tk := agg.GetTicker()
	assert.Equal(t, "BTC_USDT", tk.PairID)
	assert.True(t, tk.LastPrice.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, tk.Open.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, tk.Close.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, tk.High.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, tk.Low.Equal(decimal.NewFromFloat(50000)))
	assert.True(t, tk.Volume.Equal(decimal.NewFromFloat(1)))
}

func TestAggregator_AddTrade_MultiplePrices(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	now := time.Now()

	agg.AddTrade(decimal.NewFromFloat(50000), decimal.NewFromFloat(1), now)
	agg.AddTrade(decimal.NewFromFloat(51000), decimal.NewFromFloat(0.5), now)
	agg.AddTrade(decimal.NewFromFloat(49000), decimal.NewFromFloat(2), now)

	tk := agg.GetTicker()
	assert.True(t, tk.High.Equal(decimal.NewFromFloat(51000)))
	assert.True(t, tk.Low.Equal(decimal.NewFromFloat(49000)))
	assert.True(t, tk.Close.Equal(decimal.NewFromFloat(49000)))
	assert.True(t, tk.Volume.Equal(decimal.NewFromFloat(3.5)))
}

func TestAggregator_AddTrade_DifferentMinutes(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	now := time.Now()

	agg.AddTrade(decimal.NewFromFloat(50000), decimal.NewFromFloat(1), now.Add(-2*time.Minute))
	agg.AddTrade(decimal.NewFromFloat(51000), decimal.NewFromFloat(1), now.Add(-1*time.Minute))
	agg.AddTrade(decimal.NewFromFloat(52000), decimal.NewFromFloat(1), now)

	tk := agg.GetTicker()
	assert.True(t, tk.High.Equal(decimal.NewFromFloat(52000)))
	assert.True(t, tk.Volume.Equal(decimal.NewFromFloat(3)))
	assert.True(t, tk.Close.Equal(decimal.NewFromFloat(52000)))
}

func TestAggregator_GetTicker_Empty(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	tk := agg.GetTicker()

	assert.Equal(t, "BTC_USDT", tk.PairID)
	assert.True(t, tk.LastPrice.IsZero())
	assert.True(t, tk.Volume.IsZero())
}

func TestAggregator_ChangePct(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	now := time.Now()

	agg.AddTrade(decimal.NewFromFloat(50000), decimal.NewFromFloat(1), now.Add(-1*time.Minute))
	agg.AddTrade(decimal.NewFromFloat(55000), decimal.NewFromFloat(1), now)

	tk := agg.GetTicker()
	// change = (55000 - 50000) / 50000 * 100 = 10%
	assert.True(t, tk.ChangePct.Equal(decimal.NewFromFloat(10)))
}

func TestAggregator_Concurrent(t *testing.T) {
	agg := NewAggregator("BTC_USDT")
	now := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			price := decimal.NewFromFloat(50000 + float64(idx))
			agg.AddTrade(price, decimal.NewFromFloat(0.01), now)
		}(i)
		go func() {
			defer wg.Done()
			agg.GetTicker()
		}()
	}
	wg.Wait()

	tk := agg.GetTicker()
	assert.True(t, tk.Volume.GreaterThan(decimal.Zero))
}
