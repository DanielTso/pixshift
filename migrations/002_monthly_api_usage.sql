CREATE TABLE IF NOT EXISTS monthly_api_usage (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month       DATE NOT NULL,
    count       INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, month)
);
