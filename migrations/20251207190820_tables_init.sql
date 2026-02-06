-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_url TEXT UNIQUE NOT NULL,
    clicks INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Seeding with some sample data.
-- This one will be deleted after first cron execution and is invalid from beginning.
INSERT INTO urls (id, original_url, short_url, clicks, created_at, valid_until)
    VALUES ( gen_random_uuid(), 'https://github.com', 'GITHUB', 0, NOW(), NOW());

INSERT INTO urls (id, original_url, short_url, clicks, created_at, valid_until)
    VALUES (gen_random_uuid(), 'https://google.com', 'GOOGLE', 0, NOW(), DATE('2029-01-01'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
