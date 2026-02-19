-- Move type and features from Revision to Feature level.
-- 1. Add type to features (default from latest revision per feature).
ALTER TABLE features ADD COLUMN type TEXT NOT NULL DEFAULT 'annotation';

-- 2. Backfill feature type from latest revision per feature.
UPDATE features SET type = (
    SELECT fr.type FROM feature_revisions fr
    WHERE fr.dataset_id = features.dataset_id AND fr.feature_id = features.id
    ORDER BY fr.updated_at DESC LIMIT 1
) WHERE EXISTS (SELECT 1 FROM feature_revisions fr WHERE fr.dataset_id = features.dataset_id AND fr.feature_id = features.id);

-- 3. Create feature_features (parent feature -> child feature IDs; replaces feature_revision_features).
CREATE TABLE IF NOT EXISTS feature_features (
    dataset_id          TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_id          TEXT NOT NULL,
    child_feature_id    TEXT NOT NULL,
    sort_order          INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (dataset_id, feature_id, child_feature_id),
    FOREIGN KEY (dataset_id, feature_id) REFERENCES features(dataset_id, id) ON DELETE CASCADE,
    FOREIGN KEY (dataset_id, child_feature_id) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);

-- 4. Migrate feature_revision_features into feature_features using latest revision per feature.
INSERT INTO feature_features (dataset_id, feature_id, child_feature_id, sort_order)
SELECT frf.dataset_id, fr.feature_id, frf.feature_id, frf.sort_order
FROM feature_revision_features frf
JOIN feature_revisions fr ON fr.dataset_id = frf.dataset_id AND fr.id = frf.feature_revision_id
JOIN (
    SELECT dataset_id, feature_id, MAX(updated_at) AS max_updated
    FROM feature_revisions
    GROUP BY dataset_id, feature_id
) latest ON fr.dataset_id = latest.dataset_id AND fr.feature_id = latest.feature_id AND fr.updated_at = latest.max_updated;

-- 5. Drop type from feature_revisions by recreating the table.
CREATE TABLE feature_revisions_new (
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
    PRIMARY KEY (dataset_id, id),
    FOREIGN KEY (dataset_id, feature_id) REFERENCES features(dataset_id, id) ON DELETE CASCADE
);
INSERT INTO feature_revisions_new (id, dataset_id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note)
SELECT id, dataset_id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note
FROM feature_revisions;
DROP TABLE feature_revisions;
ALTER TABLE feature_revisions_new RENAME TO feature_revisions;
CREATE INDEX IF NOT EXISTS idx_feature_revisions_feature_id ON feature_revisions(feature_id);
CREATE INDEX IF NOT EXISTS idx_feature_revisions_updated_at ON feature_revisions(updated_at);

-- 6. Drop old revision-level features table.
DROP TABLE feature_revision_features;
