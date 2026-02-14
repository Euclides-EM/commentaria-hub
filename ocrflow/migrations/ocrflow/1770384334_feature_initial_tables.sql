-- Feature (ocrflow.Meta + IsRoot, IsDefault)
CREATE TABLE IF NOT EXISTS features (
    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    id            TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name          TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    is_root       INTEGER NOT NULL DEFAULT 0,
    is_default    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (dataset_id, id)
);

CREATE INDEX IF NOT EXISTS idx_features_updated_at ON features(updated_at);
CREATE INDEX IF NOT EXISTS idx_features_dataset_id ON features(dataset_id);

-- FeatureRevision (ocrflow.Meta + Prompt, Regex, ExecutionStrategy, Note, Type; parent feature)
CREATE TABLE IF NOT EXISTS feature_revisions (
    id                   TEXT NOT NULL,
    dataset_id           TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_id           TEXT NOT NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name                 TEXT NOT NULL DEFAULT '',
    description          TEXT NOT NULL DEFAULT '',
    prompt               TEXT NOT NULL DEFAULT '',
    regex                TEXT NOT NULL DEFAULT '',
    execution_strategy   TEXT NOT NULL,
    note                 TEXT NOT NULL DEFAULT '',
    type                 TEXT NOT NULL,
    PRIMARY KEY (dataset_id, id),
    FOREIGN KEY (dataset_id, feature_id) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_feature_revisions_feature_id ON feature_revisions(feature_id);
CREATE INDEX IF NOT EXISTS idx_feature_revisions_updated_at ON feature_revisions(updated_at);

-- FeatureRevision.Features (list of feature IDs referenced by this revision)
CREATE TABLE IF NOT EXISTS feature_revision_features (
    dataset_id           TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_revision_id  TEXT NOT NULL,
    feature_id           TEXT NOT NULL,
    sort_order           INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (dataset_id, feature_revision_id, feature_id),
    FOREIGN KEY (dataset_id, feature_revision_id) REFERENCES feature_revisions(dataset_id, id) ON DELETE CASCADE,
    FOREIGN KEY (dataset_id, feature_id) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);

-- FeatureExecution (ocrflow.Meta + Collection, Keys, Policy, Status)
CREATE TABLE IF NOT EXISTS feature_executions (
    id             TEXT PRIMARY KEY NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name           TEXT NOT NULL DEFAULT '',
    description    TEXT NOT NULL DEFAULT '',
    dataset_id     TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    annotation_id   TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    keys           TEXT NOT NULL DEFAULT '[]',
    policy_skip_if TEXT NOT NULL DEFAULT '',
    status         TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_feature_executions_status ON feature_executions(status);
CREATE INDEX IF NOT EXISTS idx_feature_executions_updated_at ON feature_executions(updated_at);

-- FeatureExecution.Apply (feature + revision per item)
CREATE TABLE IF NOT EXISTS feature_execution_apply (
    execution_id  TEXT NOT NULL,
    dataset_id    TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    annotation_id TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    feature_id    TEXT NOT NULL,
    revision_id   TEXT NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (execution_id, sort_order),
    FOREIGN KEY (execution_id) REFERENCES feature_executions(id) ON DELETE CASCADE,
    FOREIGN KEY (dataset_id, feature_id) REFERENCES features(dataset_id, id),
    FOREIGN KEY (dataset_id, revision_id) REFERENCES feature_revisions(dataset_id, id)
);

-- FeatureResult (Feature, Key, Source, Values, Note; no Meta)
-- Source and Values stored as JSON (Values is recursive)
CREATE TABLE IF NOT EXISTS feature_results (
    dataset_id           TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature       TEXT NOT NULL,
    key           TEXT NOT NULL,
    note          TEXT NOT NULL DEFAULT '',
    source_resp   TEXT NOT NULL DEFAULT '',
    source_id     TEXT NOT NULL DEFAULT '',
    source_revision TEXT NOT NULL DEFAULT '',
    source_name   TEXT NOT NULL DEFAULT '',
    values_json   TEXT NOT NULL DEFAULT '[]',
    PRIMARY KEY (dataset_id, feature, key),
    FOREIGN KEY (dataset_id, feature) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_feature_results_feature ON feature_results(feature);
CREATE INDEX IF NOT EXISTS idx_feature_results_dataset_id ON feature_results(dataset_id);
