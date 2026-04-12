package engine

import (
	"context"
	"fmt"
)

// Engine runs a matching loop for a single trading pair in a dedicated goroutine.
type Engine struct {
	book   *OrderBook
	cmdCh  chan Command
	pairID string
}

func NewEngine(pairID string, bufferSize int) *Engine {
	return &Engine{
		book:   NewOrderBook(pairID),
		cmdCh:  make(chan Command, bufferSize),
		pairID: pairID,
	}
}

// Start begins the matching loop. Blocks until ctx is cancelled or CmdStop received.
func (e *Engine) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case cmd := <-e.cmdCh:
				switch cmd.Type {
				case CmdSubmit:
					trades := e.book.Match(cmd.Order)
					if cmd.Order.Remaining.IsPositive() {
						e.book.Add(cmd.Order)
					}
					cmd.ResultCh <- Result{Trades: trades, OrderID: cmd.Order.ID}
				case CmdCancel:
					err := e.book.Cancel(cmd.OrderID)
					cmd.ResultCh <- Result{Err: err, OrderID: cmd.OrderID}
				case CmdStop:
					cmd.ResultCh <- Result{}
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Submit sends an order to the engine and waits for the result.
func (e *Engine) Submit(order *BookOrder) ([]*Trade, error) {
	resultCh := make(chan Result, 1)
	select {
	case e.cmdCh <- Command{Type: CmdSubmit, Order: order, ResultCh: resultCh}:
	default:
		return nil, fmt.Errorf("engine %s: command buffer full", e.pairID)
	}
	res := <-resultCh
	return res.Trades, res.Err
}

// Cancel removes an order from the book.
func (e *Engine) Cancel(orderID string) error {
	resultCh := make(chan Result, 1)
	select {
	case e.cmdCh <- Command{Type: CmdCancel, OrderID: orderID, ResultCh: resultCh}:
	default:
		return fmt.Errorf("engine %s: command buffer full", e.pairID)
	}
	res := <-resultCh
	return res.Err
}

// Stop gracefully stops the engine goroutine.
func (e *Engine) Stop() {
	resultCh := make(chan Result, 1)
	e.cmdCh <- Command{Type: CmdStop, ResultCh: resultCh}
	<-resultCh
}

// Preload inserts orders into the book without triggering matching (crash recovery).
func (e *Engine) Preload(orders []*BookOrder) {
	for _, order := range orders {
		e.book.Add(order)
	}
}

// OrderCount returns current number of resting orders.
func (e *Engine) OrderCount() int {
	return e.book.OrderCount()
}
