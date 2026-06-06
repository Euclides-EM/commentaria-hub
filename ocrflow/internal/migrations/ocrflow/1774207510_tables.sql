CREATE TABLE IF NOT EXISTS datasets
(
    id             TEXT PRIMARY KEY NOT NULL,
    name           VARCHAR(255)     NOT NULL,
    description    TEXT                      DEFAULT '' not null,
    created_at     TIMESTAMP                 DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP                 DEFAULT CURRENT_TIMESTAMP NOT NULL,
    edition_id     TEXT             NOT NULL,
    facsimile_id   TEXT             NOT NULL,
    dpi            INTEGER          NOT NULL,
    deskewed       BOOLEAN          NOT NULL DEFAULT FALSE,
    status         TEXT             NOT NULL DEFAULT 'ready',
    creation_error TEXT             NOT NULL DEFAULT '',
    pages          TEXT             NOT NULL DEFAULT '',
    denoised       BOOLEAN          NOT NULL DEFAULT FALSE,
    FOREIGN KEY (edition_id, facsimile_id) REFERENCES facsimiles ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS annotations
(
    id                   TEXT PRIMARY KEY NOT NULL,
    name                 VARCHAR(255)     NOT NULL,
    description          TEXT                      DEFAULT '' NOT NULL,
    created_at           TIMESTAMP        NOT NULL,
    updated_at           TIMESTAMP        NOT NULL,
    pages                TEXT                      DEFAULT '' NOT NULL,
    segmented            BOOLEAN          NOT NULL DEFAULT FALSE,
    ground_truth         BOOLEAN          NOT NULL DEFAULT FALSE,
    ocred                BOOLEAN          NOT NULL DEFAULT FALSE,
    dataset_id           TEXT             NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    origin_annotation_id TEXT                      DEFAULT '' NOT NULL,
    hidden               BOOLEAN          NOT NULL DEFAULT FALSE,
    lines_detected       BOOLEAN          NOT NULL DEFAULT FALSE
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
CREATE TABLE IF NOT EXISTS models
(
    id               TEXT PRIMARY KEY                       NOT NULL,
    name             VARCHAR(255)                           NOT NULL,
    description      TEXT                        DEFAULT '' NOT NULL,
    created_at       TIMESTAMP                              NOT NULL,
    updated_at       TIMESTAMP                              NOT NULL,
    type             TEXT                                   NOT NULL,
    location         TEXT                                   NOT NULL,
    algorithm_family TEXT                        DEFAULT '' NOT NULL,
    local_path       TEXT                        DEFAULT '' NOT NULL,
    base_model_id    TEXT REFERENCES models (id) DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS model_categories
(
    model_id TEXT NOT NULL,
    category TEXT NOT NULL,
    PRIMARY KEY (model_id, category),
    FOREIGN KEY (model_id) REFERENCES models (id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS models_base_annotations
(
    model_id      TEXT NOT NULL REFERENCES models (id) ON DELETE CASCADE,
    dataset_id    TEXT NOT NULL REFERENCES datasets (id),
    annotation_id TEXT NOT NULL REFERENCES annotations (id),
    PRIMARY KEY (model_id, dataset_id, annotation_id)
);
CREATE TABLE IF NOT EXISTS annotation_applied_rules
(
    id            INTEGER PRIMARY KEY,
    annotation_id TEXT    NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    rule_id       TEXT    NOT NULL REFERENCES annotation_rules (id) ON DELETE CASCADE,
    applied_index INTEGER NOT NULL,
    applied_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE (annotation_id, applied_index)
);
CREATE TABLE IF NOT EXISTS "facsimiles"
(
    edition_id      TEXT            NOT NULL,
    id              TEXT            NOT NULL,
    created_at      TIMESTAMP       NOT NULL,
    updated_at      TIMESTAMP       NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT DEFAULT '' NOT NULL,
    url             TEXT            NOT NULL,
    main_text_pages TEXT DEFAULT '' NOT NULL,
    PRIMARY KEY (edition_id, id)
);
CREATE TABLE IF NOT EXISTS features
(
    id          TEXT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT                  DEFAULT '' NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    dataset_id  TEXT REFERENCES datasets (id) ON DELETE CASCADE,
    scope       TEXT         NOT NULL DEFAULT 'dataset',

    is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_list     BOOLEAN      NOT NULL DEFAULT FALSE,
    color                    NOT NULL DEFAULT '#000000' CHECK (color <> ''),

    -- JSON array of strings
    properties  TEXT         NOT NULL DEFAULT '[]'
);
CREATE TABLE IF NOT EXISTS feature_revisions
(
    id          TEXT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT                  DEFAULT '' NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    dataset_id  TEXT REFERENCES datasets (id) ON DELETE CASCADE,
    scope       TEXT         NOT NULL DEFAULT 'dataset',
    feature_id  TEXT         NOT NULL REFERENCES features (id) ON DELETE CASCADE,

    prompt      TEXT         NOT NULL,
    categorizer TEXT         NOT NULL
);

CREATE TABLE IF NOT EXISTS feature_results
(
    name            VARCHAR(255)                        NOT NULL,
    description     TEXT      DEFAULT ''                NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    dataset_id      TEXT                                NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    feature_id      TEXT                                NOT NULL REFERENCES features (id) ON DELETE CASCADE,
    annotation_id   TEXT                                NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    page_key        TEXT                                NOT NULL,

    source_resp     TEXT                                NOT NULL,
    source_id       TEXT,
    source_revision TEXT,
    source_name     TEXT,
    PRIMARY KEY (dataset_id, feature_id, annotation_id, page_key)
);

CREATE TABLE IF NOT EXISTS feature_result_values
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    dataset_id    TEXT NOT NULL,
    feature_id    TEXT NOT NULL,
    annotation_id TEXT NOT NULL,
    page_key      TEXT NOT NULL,
    surface       TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (dataset_id, feature_id, annotation_id, page_key) REFERENCES feature_results (dataset_id, feature_id, annotation_id, page_key) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS annotation_groups
(
    id          TEXT PRIMARY KEY,
    name        TEXT               DEFAULT '',
    description TEXT               DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS annotation_group_annotations
(
    group_id      TEXT NOT NULL REFERENCES annotation_groups (id) ON DELETE CASCADE,
    dataset_id    TEXT NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    annotation_id TEXT NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, dataset_id, annotation_id)
);
CREATE TABLE IF NOT EXISTS editions_preferred_annotation
(
    edition_id    TEXT NOT NULL,
    dataset_id    TEXT NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    annotation_id TEXT NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    PRIMARY KEY (edition_id, dataset_id, annotation_id)
);
CREATE TABLE IF NOT EXISTS annotation_merged_annotations
(
    annotation_id        TEXT      NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    merged_dataset_id    TEXT      NOT NULL,
    merged_annotation_id TEXT      NOT NULL,
    merged_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (annotation_id, merged_dataset_id, merged_annotation_id, merged_at)
);

CREATE INDEX IF NOT EXISTS idx_annotation_applied_rules_ann_order ON annotation_applied_rules (annotation_id, applied_index);
CREATE INDEX IF NOT EXISTS idx_features_dataset ON features (dataset_id);
CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions (feature_id);
CREATE INDEX IF NOT EXISTS idx_results_lookup ON feature_results (dataset_id, annotation_id, feature_id, page_key);
CREATE INDEX IF NOT EXISTS idx_group_annotations_group_id
    ON annotation_group_annotations (group_id);
CREATE INDEX IF NOT EXISTS idx_group_annotations_annotation
    ON annotation_group_annotations (dataset_id, annotation_id);
CREATE INDEX IF NOT EXISTS idx_annotation_merged_annotations_annotation_id
    ON annotation_merged_annotations (annotation_id);
