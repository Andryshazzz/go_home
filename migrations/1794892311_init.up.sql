CREATE TABLE users (
    id            UUID PRIMARY KEY,
    login         VARCHAR(64)  NOT NULL,
    email         VARCHAR(255) NOT NULL,
    age           INT          NOT NULL,
    password_hash TEXT         NOT NULL,
    bio           VARCHAR(1000) NOT NULL DEFAULT '',
    deleted_at    TIMESTAMPTZ   NULL
);

CREATE UNIQUE INDEX users_login_key ON users (login);
CREATE UNIQUE INDEX users_email_key ON users (lower(email));

CREATE TABLE refresh_tokens (
    session_id UUID        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash CHAR(64)    NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);