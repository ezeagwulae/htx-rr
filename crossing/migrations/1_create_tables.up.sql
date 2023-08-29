CREATE TABLE checks
(
    id          BIGSERIAL PRIMARY KEY,
    crossing_id BIGINT                   NOT NULL,
    status      TEXT                     NOT NULL,
    checked_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE crossings
(
    id        BIGSERIAL PRIMARY KEY,
    name      TEXT             NOT NULL,
    latitude  DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL
);