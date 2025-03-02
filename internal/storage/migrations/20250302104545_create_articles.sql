-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles
(
    id          SERIAL PRIMARY KEY,
    source_id   INT NOT NULL,
    title    VARCHAR(255) NOT NULL,
    link    VARCHAR(255) NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    posted_at TIMESTAMP,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_articles_source_id
        FOREIGN KEY (source_id)
            REFERENCES sources (id)
            ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
