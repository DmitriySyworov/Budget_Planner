-- +goose Up
CREATE TABLE users(
 created_at date,
 updated_at date,
 deleted_at  date NULL,
 name varchar(64) NOT NULL,
 email varchar(64) UNIQUE NOT NULL,
 password char(36)  UNIQUE NOT NULL,
 user_uuid uuid PRIMARY KEY
);
CREATE INDEX idx_user_uuid ON users(user_uuid);
-- +goose Down
DROP TABLE IF EXISTS users