CREATE TABLE integrations (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    provider VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
