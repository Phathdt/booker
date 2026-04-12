-- name: GetTradingPair :one
SELECT * FROM trading_pairs WHERE id = $1;

-- name: CreateOrder :one
INSERT INTO orders (user_id, pair_id, side, type, price, quantity)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByIDAndUser :one
SELECT * FROM orders WHERE id = $1 AND user_id = $2;

-- name: ListOrders :many
SELECT * FROM orders
WHERE user_id = $1
  AND (sqlc.narg('pair_id')::VARCHAR IS NULL OR pair_id = sqlc.narg('pair_id'))
  AND (sqlc.narg('status')::VARCHAR IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CancelOrder :one
UPDATE orders SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND status IN ('new', 'partial')
RETURNING *;

-- name: UpdateOrderFilledQty :one
UPDATE orders SET filled_qty = $2, status = $3, updated_at = NOW()
WHERE id = $1 AND status IN ('new', 'partial') AND $2 <= quantity AND $2 > filled_qty
RETURNING *;
