CREATE INDEX IF NOT EXISTS idx_edition_feature_values_search
ON edition_feature_result_values (scope, feature_id, surface COLLATE NOCASE, edition_id);
