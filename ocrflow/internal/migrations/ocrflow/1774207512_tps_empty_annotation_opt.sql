INSERT INTO facsimiles
(edition_id, id, created_at, updated_at, name, description, url, main_text_pages)
VALUES
    ('', 'tps_facsimiles', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'Title Pages Facsimiles', 'Facsimiles for the title pages dataset', '', '')
ON CONFLICT(edition_id, id) DO NOTHING;

INSERT INTO datasets
(id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed, status, creation_error, pages, denoised)
VALUES
    ('tps', 'Title Pages', 'Title pages dataset', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '', 'tps_facsimiles', -1, 0, 'ready', '', '', 0)
ON CONFLICT(id) DO NOTHING;

INSERT INTO annotations
(id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id, hidden, lines_detected)
VALUES
    ('ann_1', 'Title Page Annotation', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '', 0, 0, 1, 'tps', '', 0, 0)
ON CONFLICT(id) DO NOTHING;