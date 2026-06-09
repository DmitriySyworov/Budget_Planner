-- +goose Up
CREATE TABLE users(
    created_at date,
    updated_at date,
    deleted_at  date NULL,
    name varchar(64) NOT NULL,
    email varchar(64) UNIQUE NOT NULL,
    password char(36)  UNIQUE NOT NULL,
    user_uuid char(36) PRIMARY KEY
);
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE budget(
    created_at date,
    updated_at date,
    deleted_at  date NULL,
    amount      numeric(15, 2) NOT NULL DEFAULT  0.0,
    start       date NOT NULL,
    finish      date NOT NULL,
    description text,
    budget_uuid char(36) PRIMARY KEY,
    user_uuid  char(36)  REFERENCES users(user_uuid) ON DELETE CASCADE
    CONSTRAINT check_date CHECK (start < finish),
    CONSTRAINT no_overlapping_budgets EXCLUDE USING gist (
        user_uuid WITH =,
        DATERANGE(start, finish, '[]') WITH &&
        )
);
CREATE TABLE expense(
    created_at date,
    updated_at date,
    deleted_at  date NULL,
    health numeric(15, 2) DEFAULT  0.0,
	sport numeric(15, 2) DEFAULT  0.0,
	supermarket numeric(15, 2) DEFAULT  0.0,
    restaurant numeric(15, 2) DEFAULT  0.0,
    leisure numeric(15, 2) DEFAULT  0.0,
    investments numeric(15, 2) DEFAULT  0.0,
    savings numeric(15, 2) DEFAULT  0.0,
    other numeric(15, 2) DEFAULT  0.0,
    budget_uuid char(36) REFERENCES  budget(budget_uuid) ON DELETE CASCADE,
    user_uuid  char(36)  REFERENCES users(user_uuid) ON DELETE CASCADE
);
CREATE INDEX idx_user_uuid ON users(user_uuid);
CREATE INDEX idx_budget_uuid ON budget(budget_uuid);
CREATE INDEX idx_budget_user_uuid ON budget(user_uuid);
CREATE INDEX idx_expense_budget_uuid ON expense(budget_uuid);
CREATE INDEX idx_expense_user_uuid ON expense(user_uuid);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_budget_deleted_at ON budget(deleted_at);
CREATE INDEX idx_expense_deleted_at ON expense(deleted_at);
-- +goose Down
DROP TABLE IF EXISTS expense;
DROP TABLE IF EXISTS budget;
DROP TABLE IF EXISTS users;