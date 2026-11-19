ALTER TABLE facsimiles ADD COLUMN shelfmark_id TEXT NOT NULL DEFAULT '';
ALTER TABLE facsimiles ADD COLUMN file_size_bytes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE facsimiles ADD COLUMN imported_at TIMESTAMP;
ALTER TABLE facsimiles ADD COLUMN facsimile_connection_confirmation_status TEXT NOT NULL DEFAULT '';
