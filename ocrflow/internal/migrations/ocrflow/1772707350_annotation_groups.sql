BEGIN;

CREATE TABLE annotation_groups (
    id TEXT PRIMARY KEY,
    name TEXT DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE annotation_group_annotations (
    group_id TEXT NOT NULL REFERENCES annotation_groups (id) ON DELETE CASCADE,
    dataset_id TEXT NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    annotation_id TEXT NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, dataset_id, annotation_id)
);

CREATE INDEX idx_group_annotations_group_id
    ON annotation_group_annotations(group_id);

CREATE INDEX idx_group_annotations_annotation
    ON annotation_group_annotations(dataset_id, annotation_id);

COMMIT;