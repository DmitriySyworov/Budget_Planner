-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE budgets(
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP DEFAULT NULL,
    amount      numeric(15, 2) NOT NULL DEFAULT  0.0,
    start       date NOT NULL,
    finish      date NOT NULL,
    description text,
    budget_uuid uuid PRIMARY KEY,
    user_uuid  uuid,
    CONSTRAINT check_date CHECK (start < finish),
    CONSTRAINT no_overlapping_budgets EXCLUDE USING gist (
        user_uuid WITH =,
        DATERANGE(start, finish, '[]') WITH &&
        )
);
CREATE TABLE expenses(
    health numeric(15, 2) DEFAULT  0.0,
	sport numeric(15, 2) DEFAULT  0.0,
	supermarket numeric(15, 2) DEFAULT  0.0,
    restaurant numeric(15, 2) DEFAULT  0.0,
    leisure numeric(15, 2) DEFAULT  0.0,
    investments numeric(15, 2) DEFAULT  0.0,
    savings numeric(15, 2) DEFAULT  0.0,
    other numeric(15, 2) DEFAULT  0.0,
    budget_uuid uuid NOT NULL UNIQUE REFERENCES budgets(budget_uuid) ON DELETE CASCADE,
    expense_uuid uuid  PRIMARY KEY
);
CREATE TYPE expense_category AS ENUM ('health', 'sport', 'supermarket', 'restaurant', 'leisure','investments', 'savings', 'other');
CREATE TABLE description_expenses (
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    category expense_category,
    expense numeric(15, 2) DEFAULT  0.0,
    description text,
    description_expense_uuid uuid PRIMARY KEY,
    expense_uuid uuid REFERENCES  expenses(expense_uuid) ON DELETE CASCADE
 );
CREATE INDEX idx_expenses ON description_expenses(expense_uuid);
CREATE INDEX idx_budget_deleted_at ON budgets(user_uuid) WHERE deleted_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS description_expense;
DROP TABLE IF EXISTS expense;
DROP TABLE IF EXISTS budget;
DROP TABLE IF EXISTS users;