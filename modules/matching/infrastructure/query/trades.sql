-- name: CreateTrade :one
INSERT INTO trades (pair_id, buy_order_id, sell_order_id, price, quantity, buyer_id, seller_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTradeByID :one
SELECT * FROM trades WHERE id = $1;

-- name: ListTradesByPair :many
SELECT * FROM trades WHERE pair_id = $1
ORDER BY executed_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOpenOrdersByPair :many
SELECT * FROM orders WHERE pair_id = $1 AND status IN ('new', 'partial')
ORDER BY created_at ASC;

-- name: GetTradingPairByID :one
SELECT * FROM trading_pairs WHERE id = $1;

-- name: ListActiveTradingPairs :many
SELECT * FROM trading_pairs WHERE status = 'active';
