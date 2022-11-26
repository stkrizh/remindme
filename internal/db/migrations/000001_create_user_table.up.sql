CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    email TEXT,
    identity TEXT,
    password_hash TEXT,
    created_at TIMESTAMP NOT NULL,
    activated_at TIMESTAMP,
    activation_token TEXT,
    last_login_at TIMESTAMP
);
CREATE UNIQUE INDEX user_email_idx ON "user" (email);
CREATE UNIQUE INDEX user_identity_idx ON "user" (identity);
CREATE UNIQUE INDEX user_activation_token_idx ON "user" (activation_token);