PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS feature_revisions_execution_ai_config_new
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

INSERT INTO feature_revisions_execution_ai_config_new (
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer
)
SELECT
    id, name, description, created_at, updated_at,
    dataset_id, scope, feature_id, prompt, categorizer
FROM feature_revisions;

DROP TABLE feature_revisions;
ALTER TABLE feature_revisions_execution_ai_config_new RENAME TO feature_revisions;

CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions (feature_id);

PRAGMA foreign_keys = ON;
