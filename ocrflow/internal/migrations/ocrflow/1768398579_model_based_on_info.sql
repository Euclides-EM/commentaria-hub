ALTER TABLE models ADD COLUMN base_model_id TEXT REFERENCES models(id) DEFAULT NULL;

CREATE TABLE IF NOT EXISTS models_base_annotations (
    model_id TEXT NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    dataset_id TEXT NOT NULL REFERENCES datasets(id),
    annotation_id TEXT NOT NULL REFERENCES annotations(id),
    PRIMARY KEY (model_id, dataset_id, annotation_id)
);
