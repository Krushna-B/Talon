CREATE TABLE IF NOT EXISTS market_snapshots (
    id            BIGSERIAL PRIMARY KEY,
    ticker        TEXT NOT NULL,
    event_ticker  TEXT,
    status        TEXT,
    yes_bid       NUMERIC,
    yes_ask       NUMERIC,
    no_bid        NUMERIC,
    no_ask        NUMERIC,
    last_price    NUMERIC,
    volume        NUMERIC,
    open_interest NUMERIC,
    close_time    TIMESTAMPTZ,
    raw           JSONB,
    captured_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_market_snapshots_ticker_time
    ON market_snapshots (ticker, captured_at DESC);