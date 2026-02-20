CREATE TABLE IF NOT EXISTS annotations
(
    id                   TEXT PRIMARY KEY NOT NULL,
    name                 VARCHAR(255)     NOT NULL,
    description          TEXT DEFAULT ''  NOT NULL,
    created_at           TIMESTAMP        NOT NULL,
    updated_at           TIMESTAMP        NOT NULL,
    pages                TEXT DEFAULT ''  NOT NULL,
    segmented            BOOLEAN          NOT NULL DEFAULT FALSE,
    ground_truth         BOOLEAN          NOT NULL DEFAULT FALSE,
    ocred                BOOLEAN          NOT NULL DEFAULT FALSE,
    dataset_id           TEXT             NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    origin_annotation_id TEXT DEFAULT ''  NOT NULL
);

CREATE TABLE IF NOT EXISTS annotation_rules
(
    id              TEXT PRIMARY KEY NOT NULL,
    name            VARCHAR(255)     NOT NULL,
    description     TEXT DEFAULT ''  NOT NULL,
    created_at      TIMESTAMP        NOT NULL,
    updated_at      TIMESTAMP        NOT NULL,
    rule_definition TEXT DEFAULT ''  NOT NULL
);

CREATE TABLE IF NOT EXISTS annotation_applied_rules (
    annotation_id TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    rule_id       TEXT NOT NULL REFERENCES annotation_rules(id) ON DELETE CASCADE,
    PRIMARY KEY (annotation_id, rule_id)
);
