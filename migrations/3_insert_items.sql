-- +goose Up
INSERT INTO items (name, description) VALUES ('Item 1', 'Description 1');
INSERT INTO items (name, description) VALUES ('Item 2', 'Description 2');

-- +goose Down
DELETE FROM items WHERE name IN ('Item 1', 'Item 2');