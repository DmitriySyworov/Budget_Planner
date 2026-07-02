--testing budget
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
('2026-05-12', '2026-05-12', null, '20232.67', '2026-08-01', '2026-09-23', 'get budget', '0f1e2d3c-4b5a-4678-9abc-def012345678', '6e5f4a3b-2c1d-4e9f-8a7b-6c5d4e3f2a1b');
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-05-12', '2026-05-12', null, '2029838.07', '2026-08-01', '2026-09-23', 'list budget', '7a5d9f3b-1c8e-4a21-9d6f-5b3c8e1a7d4f', 'c9b8a7d6-e5f4-4321-890a-bcdef1234567');
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-05-12', '2026-05-12', null, '202985678.07', '2026-10-01', '2026-11-23', 'list budget', '123e4567-e89b-41d3-a456-426614174000', 'c9b8a7d6-e5f4-4321-890a-bcdef1234567');
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-05-12', '2026-05-12', null, '678.07', '2026-10-01', '2026-11-23', 'remove budget', 'b408d27c-19d8-42c4-8675-ae92166c8cf9', '3f9b95b0-e13e-4b44-bf46-75840e8fe52a');
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '88888678.07', '2026-11-01', '2027-12-23', 'delete budget', '671127d2-15d9-43a5-956c-5266f72204d0', 'c6ccc482-9187-4baa-8925-0c60780627fe');
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '88888678.07', '2025-11-01', '2025-12-23', 'update budget', '859c7a21-dc20-410a-ba54-2c11fb6db2a8', '1b272de3-9827-4c47-8a60-2da8e80556f8');
--testing expense
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '88888678.07', '2025-11-01', '2025-12-23', 'budget for create expense', 'db0e2a3c-5014-416b-a8f2-8ef96bc752a7', '26f50b4d-389d-4de3-85f5-46bd8ecb1d03');
--update
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '88888678.07', '2025-11-01', '2025-12-23', 'budget for update expense', 'dd6049ad-8422-44af-aa6f-8de18f2f64d0', '941865fb-cde7-456b-a25e-de8159b9bd14');
INSERT INTO  expenses (health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid)  VALUES
    ('100.00', '0.00', '0.00', '0.00', '0.00', '0.00', '0.00', '0.00', 'dd6049ad-8422-44af-aa6f-8de18f2f64d0', '28df9dc1-be18-4da5-9bd9-f9c3fc88f4ef');
INSERT INTO description_expenses (created_at, expense, category, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', '100.00', 'health', 'buy pills', 'a51052ab-9cf4-4b55-b46f-c1dcdeec21a2', '28df9dc1-be18-4da5-9bd9-f9c3fc88f4ef');
--list
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '88888678.07', '2025-11-01', '2025-12-23', 'budget for list expense', '1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c', '4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9');
INSERT INTO  expenses (health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid)  VALUES
('12.03', '0.00', '0.00', '4567.00', '0.00', '12.09', '1111.00', '0.00', '1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c', '9926d83a-4be4-4298-ba98-25081b29cc36');
INSERT INTO description_expenses (created_at, category, description, description_expense_uuid, expense_uuid) VALUES
('2026-06-12', 'health', 'buy pills', '5b8b9333-d922-4a00-bf86-53d368e734bc', '9926d83a-4be4-4298-ba98-25081b29cc36');
INSERT INTO description_expenses (created_at, category, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', 'sport', 'go to the gym', '3b469e71-4777-4cf1-8c46-ea90b797b5d1', '9926d83a-4be4-4298-ba98-25081b29cc36');
--delete
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '8888.00', '2025-11-01', '2025-12-23', 'budget for delete description_expense', 'e7d4a215-ca89-4ec4-bcde-51dcbaf642bd', '6e917dcb-94ad-4cd6-b3de-857e2c943dfc');
INSERT INTO  expenses (health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid)  VALUES
    ('00.00', '1001.00', '0.00', '0.00', '0.00', '0.00', '0.00', '0.00', 'e7d4a215-ca89-4ec4-bcde-51dcbaf642bd', '82baf21b-beec-48cb-8db1-cb5d0e2c39d8');
INSERT INTO description_expenses (created_at, category, expense, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', 'sport', '1000.00', 'go to the gym', '3047d8b5-0c14-41cf-a0e2-63b7ad6ab6cd', '82baf21b-beec-48cb-8db1-cb5d0e2c39d8');
--testing finance
INSERT INTO budgets (created_at, updated_at, deleted_at, amount, start, finish, description, budget_uuid, user_uuid) VALUES
    ('2026-06-12', '2026-05-12', null, '10000.07', '2025-11-01', '2025-12-23', 'budget for check finance', '7c4e5186-b4d2-4ef8-a51b-0294cd93eb5a', 'f20a1bc3-5819-4a92-bdde-62fa9c671b2e');
INSERT INTO  expenses (health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid)  VALUES
    ('7893.00', '1244.00', '0.00', '0.00', '104.00', '0.00', '0.00', '0.00', '7c4e5186-b4d2-4ef8-a51b-0294cd93eb5a', '3b8d4e9a-7c21-4f51-863a-2a1d4f6c90bc');
INSERT INTO description_expenses (created_at, expense, category, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', '7893.00', 'health', 'buy pills', 'da927bf3-014e-4c7a-9bb5-86ef42c13d78', '3b8d4e9a-7c21-4f51-863a-2a1d4f6c90bc');
INSERT INTO description_expenses (created_at, expense, category, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', '1244.00', 'sport', 'go to the gym', '0d8563a2-ec19-42b7-a3cf-fc24e1236be4', '3b8d4e9a-7c21-4f51-863a-2a1d4f6c90bc');
INSERT INTO description_expenses (created_at, expense, category, description, description_expense_uuid, expense_uuid) VALUES
    ('2026-06-12', '104.00', 'leisure', 'go to the gym', 'a512cd4e-ef8a-4931-b02f-3d6c718a4ef2', '3b8d4e9a-7c21-4f51-863a-2a1d4f6c90bc');