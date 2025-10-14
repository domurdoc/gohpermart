CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(1024) NOT NULL,
    current NUMERIC NOT NULL CHECK(current >= 0) DEFAULT 0,
    withdrawn NUMERIC NOT NULL CHECK(withdrawn >= 0) DEFAULT 0,
    is_stale BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX unique_users_username ON users (LOWER(username));

CREATE TABLE orders (
    number VARCHAR(20) PRIMARY KEY,
    status VARCHAR(10) NOT NULL CHECK(
        status in ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')
    ),
    user_id INTEGER NOT NULL REFERENCES users (id),
    accrual NUMERIC NOT NULL CHECK(accrual >= 0) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    version INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE withdrawals (
    user_id INTEGER NOT NULL REFERENCES users (id),
    order_number VARCHAR(20) NOT NULL,
    sum NUMERIC NOT NULL CHECK(sum >= 0) DEFAULT 0,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL
);
