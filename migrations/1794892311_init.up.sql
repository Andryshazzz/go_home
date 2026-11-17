CREATE TABLE users (
    id            UUID PRIMARY KEY,
    login         VARCHAR(64)  NOT NULL,
    email         VARCHAR(255) NOT NULL,
    age           INT          NOT NULL,
    password_hash TEXT         NOT NULL,
    bio           VARCHAR(1000) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX users_login_key ON users (login);
CREATE UNIQUE INDEX users_email_key ON users (lower(email));