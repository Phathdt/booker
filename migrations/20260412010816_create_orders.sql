-- +goose Up
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id),
  pair_id VARCHAR(20) NOT NULL REFERENCES trading_pairs(id),
  side VARCHAR(4) NOT NULL,
  type VARCHAR(10) NOT NULL DEFAULT 'limit',
  price NUMERIC NOT NULL,
  quantity NUMERIC NOT NULL,
  filled_qty NUMERIC NOT NULL DEFAULT 0,
  status VARCHAR(20) NOT NULL DEFAULT 'new',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE INDEX idx_orders_pair_status ON orders(pair_id, status);

CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS orders;
