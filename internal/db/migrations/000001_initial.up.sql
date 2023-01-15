CREATE TABLE IF NOT EXISTS "user" (
    id BIGSERIAL PRIMARY KEY,
    email TEXT,
    identity TEXT,
    password_hash TEXT,
    created_at TIMESTAMP NOT NULL,
    activated_at TIMESTAMP,
    activation_token TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS user_email_idx ON "user" (email);
CREATE UNIQUE INDEX IF NOT EXISTS user_identity_idx ON "user" (identity);
CREATE UNIQUE INDEX IF NOT EXISTS user_activation_token_idx ON "user" (activation_token);


CREATE TABLE IF NOT EXISTS session (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS session_token_idx ON session (token);


CREATE TABLE IF NOT EXISTS channel (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    type TEXT NOT NULL,
    settings JSONB NOT NULL,
    verification_token TEXT,
    verified_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS channel_user_id_idx ON channel (user_id);
CREATE INDEX IF NOT EXISTS channel_type_idx ON channel (type);


CREATE TABLE IF NOT EXISTS limits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    email_channel_count INTEGER CONSTRAINT email_channel_count_positive CHECK (email_channel_count >= 0),
    telegram_channel_count INTEGER CONSTRAINT telegram_channel_count_positive CHECK (telegram_channel_count >= 0),
    active_reminder_count INTEGER CONSTRAINT active_reminder_count_positive CHECK (active_reminder_count >= 0),
    monthly_sent_reminder_count INTEGER 
        CONSTRAINT monthly_sent_reminder_count_positive CHECK (monthly_sent_reminder_count >= 0),
    reminder_every_per_day_count REAL
        CONSTRAINT reminder_every_per_day_count_positive CHECK (reminder_every_per_day_count >= 0)
);
CREATE UNIQUE INDEX IF NOT EXISTS limits_user_id_idx ON limits (user_id);


CREATE TABLE IF NOT EXISTS reminder (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    every TEXT,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    canceled_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS reminder_user_id_idx ON reminder (user_id);
CREATE INDEX IF NOT EXISTS reminder_status_idx ON reminder (status);
CREATE INDEX IF NOT EXISTS reminder_sent_at_idx ON reminder (sent_at);
CREATE INDEX IF NOT EXISTS reminder_at_idx ON reminder (at);


CREATE TABLE IF NOT EXISTS reminder_channel (
    id BIGSERIAL PRIMARY KEY,
    reminder_id BIGINT NOT NULL REFERENCES "reminder" (id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES "channel" (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS reminder_channel_reminder_id_channel_id_idx 
    ON reminder_channel (reminder_id, channel_id);