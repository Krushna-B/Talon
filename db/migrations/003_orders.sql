CREATE TABLE IF NOT EXISTS orders (
    order_id        TEXT PRIMARY KEY,        -- client-generated UUID; also the idempotency token
    venue_order_id  TEXT,                    -- id Kalshi assigns; NULL until ack
    ticker          TEXT NOT NULL,
    side            TEXT NOT NULL,
    action          TEXT NOT NULL,
    count           INTEGER NOT NULL,
    limit_price     NUMERIC NOT NULL,
    status          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT orders_side_valid     CHECK (side   IN ('yes', 'no')),
    CONSTRAINT orders_action_valid   CHECK (action IN ('buy', 'sell')),
    CONSTRAINT orders_status_valid   CHECK (status IN ('pending', 'resting', 'filled', 'canceled', 'failed')),
    CONSTRAINT orders_count_positive CHECK (count > 0)
);
