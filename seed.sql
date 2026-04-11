INSERT INTO assets (id, name, decimals) VALUES
    ('BTC',  'Bitcoin',  8),
    ('ETH',  'Ethereum', 18),
    ('USDT', 'Tether',   6)
ON CONFLICT DO NOTHING;

INSERT INTO trading_pairs (id, base_asset, quote_asset) VALUES
    ('BTC_USDT', 'BTC', 'USDT'),
    ('ETH_USDT', 'ETH', 'USDT')
ON CONFLICT DO NOTHING;
