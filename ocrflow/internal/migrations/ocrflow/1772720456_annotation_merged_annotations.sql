CREATE TABLE IF NOT EXISTS annotation_merged_annotations (
    annotation_id       TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    merged_dataset_id   TEXT NOT NULL,
    merged_annotation_id TEXT NOT NULL,
    merged_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (annotation_id, merged_dataset_id, merged_annotation_id, merged_at)
);

CREATE INDEX idx_annotation_merged_annotations_annotation_id
    ON annotation_merged_annotations(annotation_id);
