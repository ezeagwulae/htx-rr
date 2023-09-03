CREATE TABLE subscriptions
(
    id           BIGSERIAL PRIMARY KEY,
    phone_number TEXT,
    crossing_id  BIGINT,
    CONSTRAINT unique_subscription UNIQUE (phone_number, crossing_id)
);