WITH ranked_revisions AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY feature_id
            ORDER BY datetime(created_at) DESC, id DESC
        ) AS row_num
    FROM feature_revisions
)
UPDATE feature_revisions
SET
    ai_provider = 'ollama',
    ai_model = 'gpt-oss:120b'
WHERE id IN (
    SELECT id
    FROM ranked_revisions
    WHERE row_num = 1
);
