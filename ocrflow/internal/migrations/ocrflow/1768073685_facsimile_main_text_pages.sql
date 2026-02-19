ALTER TABLE facsimiles ADD COLUMN main_text_pages TEXT DEFAULT '' NOT NULL;

UPDATE facsimiles SET main_text_pages = '15-320,388-655' WHERE edition_id = 'Paris_1615';
UPDATE facsimiles SET main_text_pages = '9-110' WHERE edition_id = 'Paris_1598a';
UPDATE facsimiles SET main_text_pages = '13-374' WHERE edition_id = 'Paris_1667';
