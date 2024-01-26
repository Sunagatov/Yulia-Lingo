-- +goose Up
CREATE TABLE IF NOT EXISTS words (
                                     id SERIAL PRIMARY KEY,
                                     word VARCHAR(100),
    translate VARCHAR(100)
    )

-- +goose Down
DROP TABLE IF EXISTS words;
