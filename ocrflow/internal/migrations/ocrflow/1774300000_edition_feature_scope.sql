PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS features_new
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

    properties  TEXT         NOT NULL DEFAULT '[]'
);

INSERT INTO features_new (
    id, name, description, created_at, updated_at,
    dataset_id, scope, is_default, is_list, color, properties
)
SELECT
    id, name, description, created_at, updated_at,
    dataset_id, 'dataset', is_default, is_list, color, properties
FROM features;

DROP TABLE features;
ALTER TABLE features_new RENAME TO features;

CREATE TABLE IF NOT EXISTS feature_revisions_new
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

INSERT INTO feature_revisions_new (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer
)
SELECT
    id, name, description, created_at, updated_at,
    dataset_id, 'dataset', feature_id, prompt, categorizer
FROM feature_revisions;

DROP TABLE feature_revisions;
ALTER TABLE feature_revisions_new RENAME TO feature_revisions;

CREATE INDEX IF NOT EXISTS idx_features_dataset ON features (dataset_id);
CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions (feature_id);

PRAGMA foreign_keys = ON;
