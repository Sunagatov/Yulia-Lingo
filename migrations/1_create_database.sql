-- +goose Up
CREATE DATABASE testdb;

-- +goose Down
DROP DATABASE testdb;
