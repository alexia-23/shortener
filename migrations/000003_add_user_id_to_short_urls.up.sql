ALTER TABLE short_urls
    ADD COLUMN user_id VARCHAR(64);

CREATE INDEX short_urls_user_id_idx
    ON short_urls (user_id);