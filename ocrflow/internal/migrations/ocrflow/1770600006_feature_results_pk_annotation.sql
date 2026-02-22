-- Fix feature_results primary key to include annotation_id.
CREATE TABLE feature_results_new (
    dataset_id      TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    annotation_id   TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    feature         TEXT NOT NULL,
    key             TEXT NOT NULL,
    note            TEXT NOT NULL DEFAULT '',
    source_resp     TEXT NOT NULL DEFAULT '',
    source_id       TEXT NOT NULL DEFAULT '',
    source_revision TEXT NOT NULL DEFAULT '',
    source_name     TEXT NOT NULL DEFAULT '',
    values_json     TEXT NOT NULL DEFAULT '[]',
    PRIMARY KEY (dataset_id, annotation_id, feature, key),
    FOREIGN KEY (dataset_id, feature) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);
INSERT INTO feature_results_new SELECT dataset_id, annotation_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json FROM feature_results;
DROP TABLE feature_results;
ALTER TABLE feature_results_new RENAME TO feature_results;
CREATE INDEX IF NOT EXISTS idx_feature_results_feature ON feature_results(feature);
CREATE INDEX IF NOT EXISTS idx_feature_results_dataset_id ON feature_results(dataset_id);
