-- PRAGMA must run before BEGIN or it is a no-op (SQLite ignores it inside a transaction)
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

CREATE TABLE facsimiles_new
(
    edition_id      TEXT            NOT NULL,
    id              TEXT            NOT NULL,
    created_at      TIMESTAMP       NOT NULL,
    updated_at      TIMESTAMP       NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT DEFAULT '' NOT NULL,
    url             TEXT            NOT NULL,
    main_text_pages TEXT DEFAULT '' NOT NULL,
    PRIMARY KEY (edition_id, id)
);

INSERT INTO facsimiles_new (
    edition_id, id, created_at, updated_at,
    name, description, url, main_text_pages
)
SELECT
    edition_id, id, created_at, updated_at,
    name, description, url, main_text_pages
FROM facsimiles;

DROP TABLE facsimiles;
ALTER TABLE facsimiles_new RENAME TO facsimiles;

DROP TABLE editions;

COMMIT;
-- Re-enable FKs after leaving the transaction (PRAGMA is no-op inside a transaction)
PRAGMA foreign_keys = ON;
-- Sanity check: returns rows if any FK is violated
PRAGMA foreign_key_check;
