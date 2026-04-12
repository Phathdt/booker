package engine

import (
	"container/list"
	"errors"

	"github.com/shopspring/decimal"
)

var ErrOrderNotFound = errors.New("order not in book")

// orderNode links a BookOrder to its position in the book for O(1) cancel.
type orderNode struct {
	order   *BookOrder
	level   *priceLevel
	element *list.Element
}

// priceLevel holds a FIFO queue of orders at the same price.
type priceLevel struct {
	price  decimal.Decimal
	orders *list.List
}

// OrderBook maintains bid and ask sides for a single trading pair.
type OrderBook struct {
	pairID string
	bids   *sortedLevels // price DESC (highest first)
	asks   *sortedLevels // price ASC (lowest first)
	index  map[string]*orderNode
}

func NewOrderBook(pairID string) *OrderBook {
	return &OrderBook{
		pairID: pairID,
		bids:   newSortedLevels(true),  // descending
		asks:   newSortedLevels(false), // ascending
		index:  make(map[string]*orderNode),
	}
}

// Add inserts an order into the appropriate side without matching.
func (ob *OrderBook) Add(order *BookOrder) {
	side := ob.sideFor(order.Side)
	level := side.getOrCreate(order.Price)

	elem := level.orders.PushBack(order)
	ob.index[order.ID] = &orderNode{
		order:   order,
		level:   level,
		element: elem,
	}
}

// Cancel removes an order from the book in O(1).
func (ob *OrderBook) Cancel(orderID string) error {
	node, ok := ob.index[orderID]
	if !ok {
		return ErrOrderNotFound
	}

	node.level.orders.Remove(node.element)
	if node.level.orders.Len() == 0 {
		side := ob.sideFor(node.order.Side)
		side.remove(node.level.price)
	}
	delete(ob.index, orderID)
	return nil
}

// Match attempts to fill an incoming order against the opposite side.
// Returns trades produced. Caller must Add() remaining qty if any.
func (ob *OrderBook) Match(incoming *BookOrder) []*Trade {
	opposite := ob.oppositeFor(incoming.Side)
	var trades []*Trade

	for opposite.len() > 0 && incoming.Remaining.IsPositive() {
		bestLevel := opposite.best()
		if bestLevel == nil {
			break
		}

		// Check if prices cross
		if !pricesCross(incoming, bestLevel.price) {
			break
		}

		// Walk FIFO queue at this price level
		filled := false
		elem := bestLevel.orders.Front()
		for elem != nil && incoming.Remaining.IsPositive() {
			resting := elem.Value.(*BookOrder)
			next := elem.Next()

			// Self-trade prevention
			if resting.UserID == incoming.UserID {
				elem = next
				continue
			}
			filled = true

			fillQty := decimal.Min(incoming.Remaining, resting.Remaining)
			tradePrice := resting.Price // resting order's price (maker price)

			var trade *Trade
			if incoming.Side == SideBuy {
				trade = NewTrade(ob.pairID, incoming, resting, tradePrice, fillQty)
			} else {
				trade = NewTrade(ob.pairID, resting, incoming, tradePrice, fillQty)
			}
			trades = append(trades, trade)

			incoming.Remaining = incoming.Remaining.Sub(fillQty)
			resting.Remaining = resting.Remaining.Sub(fillQty)

			if resting.Remaining.IsZero() {
				bestLevel.orders.Remove(elem)
				delete(ob.index, resting.ID)
			}

			elem = next
		}

		// Remove empty price level
		if bestLevel.orders.Len() == 0 {
			opposite.remove(bestLevel.price)
		}

		// If no fills happened at this level (all self-trade), move to next level
		if !filled {
			break
		}
	}

	return trades
}

// BestBid returns the highest bid price, or nil if no bids.
func (ob *OrderBook) BestBid() *decimal.Decimal {
	if ob.bids.len() == 0 {
		return nil
	}
	p := ob.bids.best().price
	return &p
}

// BestAsk returns the lowest ask price, or nil if no asks.
func (ob *OrderBook) BestAsk() *decimal.Decimal {
	if ob.asks.len() == 0 {
		return nil
	}
	p := ob.asks.best().price
	return &p
}

// OrderCount returns the number of orders in the book.
func (ob *OrderBook) OrderCount() int {
	return len(ob.index)
}

func (ob *OrderBook) sideFor(s Side) *sortedLevels {
	if s == SideBuy {
		return ob.bids
	}
	return ob.asks
}

func (ob *OrderBook) oppositeFor(s Side) *sortedLevels {
	if s == SideBuy {
		return ob.asks
	}
	return ob.bids
}

func pricesCross(incoming *BookOrder, restingPrice decimal.Decimal) bool {
	if incoming.Side == SideBuy {
		return incoming.Price.GreaterThanOrEqual(restingPrice)
	}
	return incoming.Price.LessThanOrEqual(restingPrice)
}

// sortedLevels wraps treemap for price-level management.
type sortedLevels struct {
	levels map[string]*priceLevel // price.String() → level
	sorted []*priceLevel          // maintained sorted
	desc   bool                   // true = bids (descending)
}

func newSortedLevels(desc bool) *sortedLevels {
	return &sortedLevels{
		levels: make(map[string]*priceLevel),
		desc:   desc,
	}
}

func (sl *sortedLevels) len() int {
	return len(sl.sorted)
}

func (sl *sortedLevels) best() *priceLevel {
	if len(sl.sorted) == 0 {
		return nil
	}
	return sl.sorted[0]
}

func (sl *sortedLevels) getOrCreate(price decimal.Decimal) *priceLevel {
	key := price.String()
	if lvl, ok := sl.levels[key]; ok {
		return lvl
	}

	lvl := &priceLevel{
		price:  price,
		orders: list.New(),
	}
	sl.levels[key] = lvl
	sl.insert(lvl)
	return lvl
}

func (sl *sortedLevels) remove(price decimal.Decimal) {
	key := price.String()
	delete(sl.levels, key)

	for i, lvl := range sl.sorted {
		if lvl.price.Equal(price) {
			sl.sorted = append(sl.sorted[:i], sl.sorted[i+1:]...)
			return
		}
	}
}

func (sl *sortedLevels) insert(lvl *priceLevel) {
	pos := 0
	for pos < len(sl.sorted) {
		cmp := lvl.price.Cmp(sl.sorted[pos].price)
		if sl.desc {
			if cmp > 0 {
				break
			}
		} else {
			if cmp < 0 {
				break
			}
		}
		pos++
	}

	sl.sorted = append(sl.sorted, nil)
	copy(sl.sorted[pos+1:], sl.sorted[pos:])
	sl.sorted[pos] = lvl
}
