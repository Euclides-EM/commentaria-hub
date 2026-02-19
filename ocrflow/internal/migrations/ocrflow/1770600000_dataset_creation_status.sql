-- Add status and creation_error for async dataset creation.
ALTER TABLE datasets ADD COLUMN status TEXT NOT NULL DEFAULT 'ready';
ALTER TABLE datasets ADD COLUMN creation_error TEXT NOT NULL DEFAULT '';
