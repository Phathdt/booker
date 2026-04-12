-- +goose Up
CREATE TABLE trades (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  pair_id VARCHAR(20) NOT NULL REFERENCES trading_pairs(id),
  buy_order_id UUID NOT NULL REFERENCES orders(id),
  sell_order_id UUID NOT NULL REFERENCES orders(id),
  price NUMERIC NOT NULL,
  quantity NUMERIC NOT NULL,
  buyer_id UUID NOT NULL REFERENCES users(id),
  seller_id UUID NOT NULL REFERENCES users(id),
  executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_pair_executed ON trades(pair_id, executed_at DESC);
CREATE INDEX idx_trades_buyer ON trades(buyer_id, executed_at DESC);
CREATE INDEX idx_trades_seller ON trades(seller_id, executed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS trades;
