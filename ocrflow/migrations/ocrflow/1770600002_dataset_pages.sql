-- Optional page range for dataset (e.g. "1,3-5,10"); empty means all pages.
ALTER TABLE datasets ADD COLUMN pages TEXT NOT NULL DEFAULT '';
