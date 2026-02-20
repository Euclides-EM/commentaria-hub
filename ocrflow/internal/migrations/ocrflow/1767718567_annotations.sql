CREATE TABLE IF NOT EXISTS annotations
(
    id                   TEXT PRIMARY KEY NOT NULL,
    name                 VARCHAR(255)     NOT NULL,
    description          TEXT DEFAULT ''  NOT NULL,
    created_at           TIMESTAMP        NOT NULL,
    updated_at           TIMESTAMP        NOT NULL,
    pages                TEXT DEFAULT ''  NOT NULL,
    segmented            BOOLEAN          NOT NULL DEFAULT FALSE,
    ground_truth         BOOLEAN          NOT NULL DEFAULT FALSE,
    ocred                BOOLEAN          NOT NULL DEFAULT FALSE,
    dataset_id           TEXT             NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    origin_annotation_id TEXT DEFAULT ''  NOT NULL REFERENCES annotations (id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS annotation_rules
(
    id              TEXT PRIMARY KEY NOT NULL,
    name            VARCHAR(255)     NOT NULL,
    description     TEXT DEFAULT ''  NOT NULL,
    created_at      TIMESTAMP        NOT NULL,
    updated_at      TIMESTAMP        NOT NULL,
    rule_definition TEXT DEFAULT ''  NOT NULL
);

CREATE TABLE IF NOT EXISTS annotation_applied_rules (
    annotation_id TEXT NOT NULL REFERENCES annotations(id) ON DELETE CASCADE,
    rule_id       TEXT NOT NULL REFERENCES annotation_rules(id) ON DELETE CASCADE,
    PRIMARY KEY (annotation_id, rule_id)
);

-- Notes:
-- 1) name is set to the same value as id (change if model.Meta has a separate Name)
-- 2) created_at and updated_at use CURRENT_TIMESTAMP
-- 3) segmented=1 for segmentation-style annotations, segmented=0 for OCR/transcription-style annotations

BEGIN TRANSACTION;

--------------------------------------------------------------------------------
-- annotation_rules
--------------------------------------------------------------------------------

INSERT INTO annotation_rules (id, name, description, created_at, updated_at, rule_definition)
VALUES ('rule_segment_1615FineTunedCapricciosaM_0312',
        'Segment: 1615FineTunedCapricciosaM_0312',
        'Segment pages using model 1615FineTunedCapricciosaM_0312',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"segment","model":"1615FineTunedCapricciosaM_0312"}'),
       ('rule_segment_1598FineTuned16150312_0101',
        'Segment: 1598FineTuned16150312_0101',
        'Segment pages using model 1598FineTuned16150312_0101',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"segment","model":"1598FineTuned16150312_0101"}'),
       ('rule_slice_15_320_388_655',
        'SlicePages: 15-320,388-655',
        'Restrict processing to pages 15-320,388-655',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"slice_pages","pages":"15-320,388-655"}'),
       ('rule_slice_9_110',
        'SlicePages: 9-110',
        'Restrict processing to pages 9-110',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"slice_pages","pages":"9-110"}'),
       ('rule_remove_categories_ucxw7g',
        'RemoveCategories: ucxw7g',
        'Remove MainZone-P, MainZone-P--Italics, MainZone-P--Enunciation',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"remove_categories","categories":["MainZone-P--Italics","MainZone-P--Enunciation","MainZone-P"]}'),
       ('rule_remove_categories_9yvgi8',
        'RemoveCategories: 9yvgi8',
        'Remove MainZone-P, MainZone-P--Italics',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"remove_categories","categories":["MainZone-P--Italics","MainZone-P"]}'),
       ('rule_remove_overlap_precision_1000',
        'RemoveOverlap: precision 1000',
        'Remove overlap against non-text zones, precision 1000',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"remove_overlap","categories":["DigitizationArtefactZone","GraphicZone-Decoration","GraphicZone-Diagram","MainZone","MainZone-Head--Book","MainZone-Head--Section","NumberingZone","QuireMarksZone","RunningTitleZone"],"precision":1000}'),
       ('rule_lines_detect_mainzone',
        'LinesDetect: MainZone',
        'Detect lines in MainZone while ignoring selected categories',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        '{"type":"lines_detect","include_categories":["MainZone"],"ignore_categories":["CatchWord","DigitizationArtefactZone","DropCapitalZone","GraphicZone-Decoration","GraphicZone-Diagram","NumberingZone","QuireMarksZone","RunningTitleZone"]}');

--------------------------------------------------------------------------------
-- annotations
--------------------------------------------------------------------------------

-- Sentinel row so that origin_annotation_id = '' is a valid self-reference (no origin).
INSERT INTO annotations
(id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id)
VALUES ('', '(no origin)', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '', 0, 0, 0, 'rrpbnk', '');

INSERT INTO annotations
(id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id)
VALUES
    -- Manually annotated (segmented) ground truth datasets
    ('gog80x',
     'GT-Seg-MainOnly',
     'Manually annotated (ground truth); Only one big MainZone, without any subtype like paragraph, enunciation, etc.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609',
     1,
     1,
     0,
     'rrpbnk',
     ''),
    ('f0k3ks',
     'GT-Seg-SubtypesPoly',
     'Manually annotated (ground truth); Polygons for subtypes like paragraph, enunciation and main zones.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609',
     1,
     1,
     0,
     'rrpbnk',
     ''),
    ('idim36',
     'GT-Seg-Main+Sub',
     'Manually annotated (ground truth); Includes a big MainZone, in addition to subtypes like paragraph, enunciation, etc. In some cases, instead of boxes polygons were drawn to better fit the text areas.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '66,160,197,303,497,20,49,91,95-97,148-149,153,183,195-196,255,257,295-297,315,388,395-397,450-465,495-496,508,596,603,624,256,339,387,595,597,609',
     1,
        1,
        0,
     'rrpbnk',
     ''),
    ('ht01bz',
     'GT-Sub+DropCapPlain',
     'Manually segmented (ground truth); I applied the 1615FineTunedCapricciosaM_0312 model, then manually corrected the annotations in Roboflow. I added a new category ''DropCapitalZone-Plane'' for drop capitals.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '33-34,64,20-21,25-26,35,37,45,53,60,66,69-70,74-77,92,94,103,107,38,50,57,61,104,106',
     1,
     1,
     0,
     'mq9w7q',
     ''),

    -- Manually annotated (transcribed) ground truth dataset for OCR evaluation
    ('05awr4',
     'GT-OCR-Transcribed',
     'Manually transcribed (ground truth), using the Galiccorpor OCR model as a base. This transcription is the base of the 1615FineTunedGallicorpor_0301 OCR model.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '85,129,246-247,249,443,509,512,515,517',
     1,
     1,
     1,
     'uk5wbj',
     ''),

    -- Inferred annotations
    ('4s48pk',
     'Cap1615-Raw',
     'Inferred Segmentation annotations from the 1615FineTunedCapricciosaM_0312 model.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '15-655',
     1,
     0,
        0,
     'uk5wbj',
     ''),
    ('ucxw7g',
     'Cap1615-NoSub',
     'Inferred annotations from the 1615FineTunedCapricciosaM_0312 model without subtypes of the main zone (except headers).',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '15-320,388-655',
     1,
     0,
     0,
     'uk5wbj',
     ''),
    ('7plb84',
     'Cap1615-NoSub-OCR',
     'Inferred annotations from the 1615FineTunedCapricciosaM_0312 model without subtypes of the main zone (except headers) and fter transcribing with the 1615FineTunedGallicorpor_0301.mlmodel OCR model.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '15-320,388-655',
     1,
        0,
        1,
     'uk5wbj',
     ''),
    ('9yvgi8',
     'Cap1615-Enunc',
     'Inferred annotations from the 1615FineTunedCapricciosaM_0312 model without subtypes of the main zone (except headers AND enunciations). Very similar to the ucxw7g annotation, but here the ''enunciation'' subtype is NOT removed.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '15-320,388-655',
     1,
        0,
            0,
     'uk5wbj',
     ''),
    ('s0lik6',
     'Cap1615-1598-NoAlign',
     'Inferred annotations for 1598 Paris edition from the 1615FineTunedCapricciosaM_0312 model, no skewing applied.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '9-110',
     1,
        0,
        0,
     'nu3e82',
     ''),
    ('toq5ip',
     'Cap1615-1598-Deskew',
     'Inferred annotations for 1598 Paris edition from the 1615FineTunedCapricciosaM_0312 model, after skewing was applied.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '9-110',
     1,
        0,
        0,
     'mq9w7q',
     ''),
    ('j31d9m',
     'Cap1598-Deskew',
     'Inferred annotations for 1598 Paris edition from the 1598FineTuned16150312_0101 model, after skewing was applied.',
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP,
     '9-110',
     1,
     0,
        0,
     'mq9w7q',
     '');

--------------------------------------------------------------------------------
-- annotation_applied_rules
--------------------------------------------------------------------------------

-- 4s48pk
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('4s48pk', 'rule_segment_1615FineTunedCapricciosaM_0312');

-- ucxw7g
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('ucxw7g', 'rule_segment_1615FineTunedCapricciosaM_0312'),
       ('ucxw7g', 'rule_slice_15_320_388_655'),
       ('ucxw7g', 'rule_remove_categories_ucxw7g'),
       ('ucxw7g', 'rule_remove_overlap_precision_1000'),
       ('ucxw7g', 'rule_lines_detect_mainzone');

-- 9yvgi8
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('9yvgi8', 'rule_segment_1615FineTunedCapricciosaM_0312'),
       ('9yvgi8', 'rule_slice_15_320_388_655'),
       ('9yvgi8', 'rule_remove_categories_9yvgi8'),
       ('9yvgi8', 'rule_remove_overlap_precision_1000'),
       ('9yvgi8', 'rule_lines_detect_mainzone');

-- s0lik6
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('s0lik6', 'rule_segment_1615FineTunedCapricciosaM_0312'),
       ('s0lik6', 'rule_slice_9_110'),
       ('s0lik6', 'rule_remove_categories_ucxw7g'),
       ('s0lik6', 'rule_remove_overlap_precision_1000');

-- toq5ip
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('toq5ip', 'rule_slice_9_110'),
       ('toq5ip', 'rule_segment_1615FineTunedCapricciosaM_0312'),
       ('toq5ip', 'rule_remove_categories_ucxw7g'),
       ('toq5ip', 'rule_remove_overlap_precision_1000');

-- j31d9m
INSERT INTO annotation_applied_rules (annotation_id, rule_id)
VALUES ('j31d9m', 'rule_slice_9_110'),
       ('j31d9m', 'rule_segment_1598FineTuned16150312_0101'),
       ('j31d9m', 'rule_remove_categories_ucxw7g'),
       ('j31d9m', 'rule_remove_overlap_precision_1000');

COMMIT;
