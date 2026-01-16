ALTER TABLE models ADD COLUMN base_model_id TEXT REFERENCES models(id) DEFAULT NULL;

CREATE TABLE IF NOT EXISTS models_base_annotations (
    model_id TEXT NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    dataset_id TEXT NOT NULL REFERENCES datasets(id),
    annotation_id TEXT NOT NULL REFERENCES annotations(id),
    PRIMARY KEY (model_id, dataset_id, annotation_id)
);

-- Migrate existing models to set base_model_id and models_base_datasets
UPDATE models SET base_model_id = 'CapricciosaM' WHERE id IN (
    '1615FineTunedCapricciosaM_0312'
);
UPDATE models SET base_model_id = '1615FineTunedCapricciosaM_0312' WHERE id IN (
    '1598FineTuned16150312_0101'
);
INSERT INTO models_base_annotations (model_id, dataset_id, annotation_id) VALUES
    ('1615FineTunedCapricciosaM_0312', 'uk5wbj','ropizr'),
    ('1598FineTuned16150312_0101', 'mq9w7q', 'ht01bz'),
    ('1615FineTunedGallicorpor_0301', 'uk5wbj', '05awr4');