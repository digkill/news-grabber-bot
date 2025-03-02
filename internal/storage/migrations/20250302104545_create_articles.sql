-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles
(
    id          SERIAL PRIMARY KEY,
    source_id   INT NOT NULL,
    title    VARCHAR(255) NOT NULL,
    link    VARCHAR(255) NOT NULL,
    type_source INT(1) NOT NULL DEFAULT 0,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
