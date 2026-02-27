-- deepest children first
DROP TABLE IF EXISTS feature_result_values;
DROP TABLE IF EXISTS feature_results;
DROP TABLE IF EXISTS feature_execution_apply;
DROP TABLE IF EXISTS feature_revisions;
DROP TABLE IF EXISTS feature_features;
DROP TABLE IF EXISTS feature_executions;
DROP TABLE IF EXISTS features;

------------------------------------------------------------
-- FEATURES
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS features (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,

    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_list BOOLEAN NOT NULL DEFAULT FALSE,
    color NOT NULL DEFAULT '#000000' CHECK (color <> ''),

    -- JSON array of strings
    properties    TEXT NOT NULL DEFAULT '[]'
);

CREATE INDEX IF NOT EXISTS idx_features_dataset ON features(dataset_id);


------------------------------------------------------------
-- REVISIONS
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_revisions (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    dataset_id    TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_id    TEXT NOT NULL REFERENCES features(id) ON DELETE CASCADE,

    prompt        TEXT NOT NULL,
    categorizer   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions(feature_id);


------------------------------------------------------------
-- RESULTS (parent)
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_results (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    dataset_id    TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_id    TEXT NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    annotation_id   TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,

    page_key             TEXT NOT NULL,

    source_resp     TEXT NOT NULL,
    source_id       TEXT,
    source_revision TEXT,
    source_name     TEXT
);

CREATE INDEX IF NOT EXISTS idx_results_lookup  ON feature_results(dataset_id, annotation_id, feature_id, page_key);


------------------------------------------------------------
-- RESULT VALUES (child rows)
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS result_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    result_id    TEXT NOT NULL REFERENCES feature_results(id) ON DELETE CASCADE,
    surface      TEXT NOT NULL DEFAULT '',

    -- JSON object (map[string]string)
    properties   TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_result_values_result ON result_values(result_id);