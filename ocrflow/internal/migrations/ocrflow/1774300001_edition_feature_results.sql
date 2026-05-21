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
