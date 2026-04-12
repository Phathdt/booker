package trades

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecentTrades_Add_GetRecent(t *testing.T) {
	rt := NewRecentTrades()
	rt.Add(TradeInfo{TradeID: "t1", Price: "50000", Timestamp: 1000})
	rt.Add(TradeInfo{TradeID: "t2", Price: "51000", Timestamp: 2000})
	rt.Add(TradeInfo{TradeID: "t3", Price: "52000", Timestamp: 3000})

	result := rt.GetRecent(10)
	assert.Len(t, result, 3)
	assert.Equal(t, "t3", result[0].TradeID) // newest first
	assert.Equal(t, "t1", result[2].TradeID)
}

func TestRecentTrades_GetRecent_Limit(t *testing.T) {
	rt := NewRecentTrades()
	for i := 0; i < 10; i++ {
		rt.Add(TradeInfo{TradeID: fmt.Sprintf("t%d", i)})
	}

	result := rt.GetRecent(3)
	assert.Len(t, result, 3)
	assert.Equal(t, "t9", result[0].TradeID)
}

func TestRecentTrades_CircularBuffer_Overflow(t *testing.T) {
	rt := NewRecentTrades()
	for i := 0; i < 150; i++ {
		rt.Add(TradeInfo{TradeID: fmt.Sprintf("t%d", i)})
	}

	result := rt.GetRecent(0) // 0 = all
	assert.Len(t, result, maxTrades)
	assert.Equal(t, "t149", result[0].TradeID) // newest
	assert.Equal(t, "t50", result[99].TradeID) // oldest (0-49 evicted)
}

func TestRecentTrades_Empty(t *testing.T) {
	rt := NewRecentTrades()
	result := rt.GetRecent(10)
	assert.Empty(t, result)
}

func TestRecentTrades_Concurrent(t *testing.T) {
	rt := NewRecentTrades()
	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			rt.Add(TradeInfo{TradeID: fmt.Sprintf("t%d", idx)})
		}(i)
		go func() {
			defer wg.Done()
			rt.GetRecent(10)
		}()
	}
	wg.Wait()

	result := rt.GetRecent(0)
	assert.LessOrEqual(t, len(result), maxTrades)
}
