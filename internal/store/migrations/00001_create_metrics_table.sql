-- +goose Up
CREATE TABLE IF NOT EXISTS metrics (
   id TEXT PRIMARY KEY,
   type TEXT NOT NULL,
   delta BIGINT,
   value DOUBLE PRECISION
);

-- +goose Down
DROP TABLE IF EXISTS metrics;