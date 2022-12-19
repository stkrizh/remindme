CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    email TEXT,
    identity TEXT,
    password_hash TEXT,
    created_at TIMESTAMP NOT NULL,
    activated_at TIMESTAMP,
    activation_token TEXT
);
CREATE UNIQUE INDEX user_email_idx ON "user" (email);
CREATE UNIQUE INDEX user_identity_idx ON "user" (identity);
CREATE UNIQUE INDEX user_activation_token_idx ON "user" (activation_token);


CREATE TABLE IF NOT EXISTS session (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX session_token_idx ON session (token);


CREATE TABLE IF NOT EXISTS channel (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    settings JSONB NOT NULL,
    verification_token TEXT,
    verified_at TIMESTAMP
);
CREATE INDEX channel_user_id_idx ON channel (user_id);
CREATE INDEX channel_verification_token_idx ON channel (verification_token);