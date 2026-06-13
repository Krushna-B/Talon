CREATE TABLE IF NOT EXISTS system_state (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO system_state (key, value) VALUES
    ('TRADING_ENABLED', 'false'),
    ('MODE',            'paper'),
    ('KILL_SWITCH',     'false')
ON CONFLICT (key) DO NOTHING;
