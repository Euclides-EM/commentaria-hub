-- 20260302_change_segment_to_model_detect.sql
-- Purpose: migrate annotation_rules.rule_definition JSON so {"type":"segment", ...}
--          becomes {"type":"model_detect", ...}
BEGIN;

UPDATE annotation_rules
SET
    rule_definition = json_set(rule_definition, '$.type', 'model_detect'),
    updated_at      = CURRENT_TIMESTAMP
WHERE
    json_valid(rule_definition)
  AND json_extract(rule_definition, '$.type') = 'segment';

COMMIT;