INSERT INTO assets (id, name, decimals) VALUES
    ('BTC',  'Bitcoin',  8),
    ('ETH',  'Ethereum', 18),
    ('USDT', 'Tether',   6)
ON CONFLICT DO NOTHING;

INSERT INTO trading_pairs (id, base_asset, quote_asset) VALUES
    ('BTC_USDT', 'BTC', 'USDT'),
    ('ETH_USDT', 'ETH', 'USDT')
ON CONFLICT DO NOTHING;

-- Seed 100 test accounts with BTC, ETH, and USDT balances
-- Password: "password123" hashed with bcrypt
DO $$
DECLARE
    uid UUID;
    i INT;
BEGIN
    FOR i IN 1..100 LOOP
        INSERT INTO users (id, email, password, role, status)
        VALUES (
            uuid_generate_v4(),
            'trader' || i || '@booker.dev',
            '$2a$10$u8N1SibhTTsqSMfrn0kS5Oja46nvr60VvFW5K5SEa0wPTASuSvow6',
            'user',
            'active'
        )
        ON CONFLICT (email) DO NOTHING
        RETURNING id INTO uid;

        IF uid IS NOT NULL THEN
            INSERT INTO wallets (user_id, asset_id, available, locked) VALUES
                (uid, 'BTC',  10.0,       0),
                (uid, 'ETH',  500.0,      0),
                (uid, 'USDT', 1000000.0,  0)
            ON CONFLICT (user_id, asset_id) DO NOTHING;
        END IF;
    END LOOP;
END $$;
