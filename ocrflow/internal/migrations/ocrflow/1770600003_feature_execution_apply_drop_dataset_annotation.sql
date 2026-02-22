-- Drop redundant dataset_id and annotation_id from feature_execution_apply (they match the execution's).
CREATE TABLE feature_execution_apply_new (
    execution_id  TEXT NOT NULL,
    feature_id    TEXT NOT NULL,
    revision_id   TEXT NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (execution_id, sort_order),
    FOREIGN KEY (execution_id) REFERENCES feature_executions(id) ON DELETE CASCADE
);
INSERT INTO feature_execution_apply_new (execution_id, feature_id, revision_id, sort_order)
SELECT execution_id, feature_id, revision_id, sort_order
FROM feature_execution_apply;
DROP TABLE feature_execution_apply;
ALTER TABLE feature_execution_apply_new RENAME TO feature_execution_apply;
