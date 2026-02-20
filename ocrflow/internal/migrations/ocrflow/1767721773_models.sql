CREATE TABLE IF NOT EXISTS models
(
    id               TEXT PRIMARY KEY NOT NULL,
    name             VARCHAR(255)     NOT NULL,
    description      TEXT DEFAULT ''  NOT NULL,
    created_at       TIMESTAMP        NOT NULL,
    updated_at       TIMESTAMP        NOT NULL,
    type             TEXT             NOT NULL,
    location         TEXT             NOT NULL,
    algorithm_family TEXT DEFAULT ''  NOT NULL,
    local_path       TEXT DEFAULT ''  NOT NULL
);

CREATE TABLE IF NOT EXISTS model_categories
(
    model_id TEXT NOT NULL,
    category TEXT NOT NULL,
    PRIMARY KEY (model_id, category),
    FOREIGN KEY (model_id) REFERENCES models (id) ON DELETE CASCADE
);
