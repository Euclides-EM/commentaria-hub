-- Edition classification schema/setup.
-- Consolidated before merge: feature/revision scope support, edition-level result tables,
-- AI provider/model columns, and the m_classifier feature seed.

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

CREATE TABLE IF NOT EXISTS edition_feature_results
(
    name            VARCHAR(255)                        NOT NULL,
    description     TEXT      DEFAULT ''                NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    scope           TEXT                                NOT NULL DEFAULT 'editions',
    edition_id      TEXT                                NOT NULL,
    feature_id      TEXT                                NOT NULL REFERENCES features (id) ON DELETE CASCADE,

    source_resp     TEXT                                NOT NULL,
    source_id       TEXT,
    source_revision TEXT,
    source_name     TEXT,
    PRIMARY KEY (scope, edition_id, feature_id)
);

CREATE TABLE IF NOT EXISTS edition_feature_result_values
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    scope      TEXT NOT NULL DEFAULT 'editions',
    edition_id TEXT NOT NULL,
    feature_id TEXT NOT NULL,
    surface    TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (scope, edition_id, feature_id) REFERENCES edition_feature_results (scope, edition_id, feature_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_edition_feature_results_lookup ON edition_feature_results (scope, edition_id, feature_id);

ALTER TABLE feature_revisions
ADD COLUMN ai_provider TEXT NOT NULL DEFAULT 'openai';

ALTER TABLE feature_revisions
ADD COLUMN ai_model TEXT NOT NULL DEFAULT 'gpt-5-mini';

PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS feature_revisions_ai_config_new
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
    categorizer TEXT         NOT NULL,
    ai_provider TEXT         NOT NULL,
    ai_model    TEXT         NOT NULL
);

INSERT INTO feature_revisions_ai_config_new (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
)
SELECT
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer, ai_provider, ai_model
FROM feature_revisions;

DROP TABLE feature_revisions;
ALTER TABLE feature_revisions_ai_config_new RENAME TO feature_revisions;

CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions (feature_id);

PRAGMA foreign_keys = ON;

INSERT INTO features (
    id, name, description, created_at, updated_at,
    dataset_id, scope, is_default, is_list, color, properties
) VALUES (
    'm_classifier',
    'Subject Categories',
    'Classifies each edition against a fixed list of mathematical and related subject categories.',
    '2026-05-22T14:19:49Z',
    '2026-06-03T00:00:00Z',
    NULL,
    'editions',
    1,
    1,
    '#4E7A6A',
    '[]'
) ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    updated_at = excluded.updated_at,
    dataset_id = excluded.dataset_id,
    scope = excluded.scope,
    is_default = excluded.is_default,
    is_list = excluded.is_list,
    color = excluded.color,
    properties = excluded.properties;
