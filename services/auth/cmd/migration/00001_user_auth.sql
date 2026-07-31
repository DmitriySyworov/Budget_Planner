-- +goose Up
CREATE TABLE users(
 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
 deleted_at  TIMESTAMP DEFAULT NULL,
 name varchar(64) NOT NULL,
 email varchar(64) UNIQUE NOT NULL,
 password char(60) NOT NULL,
 user_uuid uuid PRIMARY KEY
);
CREATE INDEX idx_user_uuid ON users(user_uuid);
CREATE INDEX idx_user_deleted_at ON users(deleted_at);
-- +goose Down
DROP TABLE IF EXISTS users