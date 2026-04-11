-- +goose Up
CREATE TABLE assets (
  id VARCHAR(10) PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  decimals INT NOT NULL DEFAULT 8
);

CREATE TABLE trading_pairs (
  id VARCHAR(20) PRIMARY KEY,
  base_asset VARCHAR(10) NOT NULL REFERENCES assets(id),
  quote_asset VARCHAR(10) NOT NULL REFERENCES assets(id),
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  min_qty NUMERIC NOT NULL DEFAULT 0.00001,
  tick_size NUMERIC NOT NULL DEFAULT 0.01,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS trading_pairs;

DROP TABLE IF EXISTS assets;
