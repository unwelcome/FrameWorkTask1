CREATE TABLE users (
    uuid          UUID         PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name    VARCHAR(50)  NOT NULL,
    last_name     VARCHAR(50)  NOT NULL,
    patronymic    VARCHAR(50),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
