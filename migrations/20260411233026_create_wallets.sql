-- +goose Up
CREATE TABLE wallets (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id),
  asset_id VARCHAR(10) NOT NULL REFERENCES assets(id),
  available NUMERIC NOT NULL DEFAULT 0,
  locked NUMERIC NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, asset_id)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

-- +goose Down
DROP TABLE IF EXISTS wallets;
