ALTER TABLE annotation_applied_rules RENAME TO annotation_applied_rules_old;

CREATE TABLE annotation_applied_rules
(
    id            INTEGER PRIMARY KEY,
    annotation_id TEXT NOT NULL REFERENCES annotations (id) ON DELETE CASCADE,
    rule_id       TEXT NOT NULL REFERENCES annotation_rules (id) ON DELETE CASCADE,
    applied_index INTEGER NOT NULL,
    applied_at    TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (annotation_id, applied_index)
);

INSERT INTO annotation_applied_rules (annotation_id, rule_id, applied_index)
SELECT annotation_id,
       rule_id,
       row_number() OVER (PARTITION BY annotation_id ORDER BY rule_id) - 1
FROM annotation_applied_rules_old;

CREATE INDEX idx_annotation_applied_rules_ann_order ON annotation_applied_rules (annotation_id, applied_index);

DROP TABLE annotation_applied_rules_old;
