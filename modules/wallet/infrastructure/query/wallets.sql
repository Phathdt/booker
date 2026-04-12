-- name: GetOrCreateWallet :one
INSERT INTO wallets (user_id, asset_id)
VALUES ($1, $2)
ON CONFLICT (user_id, asset_id) DO UPDATE SET updated_at = wallets.updated_at
RETURNING *;

-- name: GetWalletByUserAndAsset :one
SELECT * FROM wallets WHERE user_id = $1 AND asset_id = $2;

-- name: GetWalletsByUserID :many
SELECT * FROM wallets WHERE user_id = $1 ORDER BY asset_id;

-- name: DepositWallet :one
UPDATE wallets
SET available = available + $3, updated_at = NOW()
WHERE user_id = $1 AND asset_id = $2
RETURNING *;

-- name: WithdrawWallet :one
UPDATE wallets
SET available = available - $3, updated_at = NOW()
WHERE user_id = $1 AND asset_id = $2 AND available >= $3
RETURNING *;

-- name: HoldWallet :one
UPDATE wallets
SET available = available - $3, locked = locked + $3, updated_at = NOW()
WHERE user_id = $1 AND asset_id = $2 AND available >= $3
RETURNING *;

-- name: ReleaseWallet :one
UPDATE wallets
SET available = available + $3, locked = locked - $3, updated_at = NOW()
WHERE user_id = $1 AND asset_id = $2 AND locked >= $3
RETURNING *;

-- name: SettleWallet :one
UPDATE wallets
SET locked = locked - $3, updated_at = NOW()
WHERE user_id = $1 AND asset_id = $2 AND locked >= $3
RETURNING *;
