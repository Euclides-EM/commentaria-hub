CREATE TABLE IF NOT EXISTS editions_preferred_annotation (
    edition_id TEXT NOT NULL REFERENCES editions(id) ON DELETE CASCADE,
    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    annotation_id TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    PRIMARY KEY (edition_id, dataset_id, annotation_id)
);