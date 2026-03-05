CREATE TABLE IF NOT EXISTS editions_preferred_annotation (
    edition_id INTEGER NOT NULL,
    dataset_id INTEGER NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    annotation_id INTEGER NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    PRIMARY KEY (edition_id, dataset_id, annotation_id)
);