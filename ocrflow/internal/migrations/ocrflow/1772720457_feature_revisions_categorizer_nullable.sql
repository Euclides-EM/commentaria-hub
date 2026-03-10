PRAGMA foreign_keys = OFF;

-- SQLite cannot alter existing column nullability or add this table-level CHECK in place,
-- so the table is rebuilt with the new schema and the data is copied over.
CREATE TABLE feature_revisions_new (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    feature_id TEXT NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    prompt TEXT,
    categorizer TEXT,
    CHECK ((NULLIF(prompt, '') IS NULL) <> (NULLIF(categorizer, '') IS NULL))
);

INSERT INTO feature_revisions_new (
    id,
    name,
    description,
    created_at,
    updated_at,
    dataset_id,
    feature_id,
    prompt,
    categorizer
)
SELECT
    id,
    name,
    description,
    created_at,
    updated_at,
    dataset_id,
    feature_id,
    CASE
        WHEN NULLIF(prompt, '') IS NOT NULL THEN prompt
        ELSE ''
    END,
    CASE
        WHEN NULLIF(prompt, '') IS NOT NULL THEN NULL
        WHEN NULLIF(categorizer, '') IS NOT NULL THEN categorizer
        ELSE ''
    END
FROM feature_revisions;

DROP TABLE feature_revisions;
ALTER TABLE feature_revisions_new RENAME TO feature_revisions;
CREATE INDEX IF NOT EXISTS idx_revisions_feature ON feature_revisions(feature_id);

PRAGMA foreign_keys = ON;
