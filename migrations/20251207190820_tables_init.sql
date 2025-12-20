-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS urls (
    original_url TEXT NOT NULL,
    short_url TEXT NOT NULL,
    clicks INTEGER NOT NULL,
    created_at_utc TIMESTAMP NOT NULL,
    PRIMARY KEY (original_url),
    UNIQUE (short_url)
);

-- Seeding with some sample data.
INSERT INTO urls (original_url, short_url, clicks, created_at_utc)
VALUES ('https://github.com', 'GITHUB', 0, NOW());

INSERT INTO urls (original_url, short_url, clicks, created_at_utc)
    VALUES ('no_url', 'BBBBBB', 0, NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
