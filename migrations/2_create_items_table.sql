-- +goose Up
CREATE TABLE IF NOT EXISTS items (
                                     id SERIAL PRIMARY KEY,
                                     name VARCHAR(100),
    description TEXT
    );

-- +goose Down
DROP TABLE IF EXISTS items;