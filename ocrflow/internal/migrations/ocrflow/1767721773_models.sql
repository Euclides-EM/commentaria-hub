CREATE TABLE IF NOT EXISTS models
(
    id               TEXT PRIMARY KEY NOT NULL,
    name             VARCHAR(255)     NOT NULL,
    description      TEXT DEFAULT ''  NOT NULL,
    created_at       TIMESTAMP        NOT NULL,
    updated_at       TIMESTAMP        NOT NULL,
    type             TEXT             NOT NULL,
    location         TEXT             NOT NULL,
    algorithm_family TEXT DEFAULT ''  NOT NULL,
    local_path       TEXT DEFAULT ''  NOT NULL
);

CREATE TABLE IF NOT EXISTS model_categories
(
    model_id TEXT NOT NULL,
    category TEXT NOT NULL,
    PRIMARY KEY (model_id, category),
    FOREIGN KEY (model_id) REFERENCES models (id) ON DELETE CASCADE
);

BEGIN TRANSACTION;

-- models
INSERT INTO models (id, name, description, created_at, updated_at, type, location, algorithm_family, local_path)
VALUES ('paris1615trained_2811', 'paris1615trained_2811', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'text', 'local', '',
        'paris1615trained_2811.pt'),
       ('CapricciosaM', 'CapricciosaM',
        'The YALTAi CapricciosaM model, without any fine tuning, downloaded from https://zenodo.org/records/10972956/files/CapricciosaM.pt',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'local', '', 'CapricciosaM.pt'),
       ('1570FineTuned_0312', '1570FineTuned_0312',
        'CapricciosaM fine tuned on 50 pages from London 1570. Annotations in https://app.roboflow.com/mia-workplace/1570-english/2',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'local', '', '1570FineTunedCapricciosaM_0312.pt'),
       ('1615FineTunedCapricciosaM_0312', '1615FineTunedCapricciosaM_0312',
        'CapricciosaM fine tuned on 50 pages from Paris 1615. Annotations in https://app.roboflow.com/mia-workplace/0212-xcfg/2',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'local', '', '1615FineTunedCapricciosaM_0312.pt'),
       ('1615FineTunedCapricciosaM_0812', '1615FineTunedCapricciosaM_0812',
        'CapricciosaM fine tuned on 50 pages from Paris 1615. It is based on pages that were automatically segmented with 1615FineTunedCapricciosaM_0312, and then multiple rules were applied on then, to fix various segmentation issues. The full specification is noted in the annotation ID ucxw7g. Then, the corrected annotations were manually corrected further in the Roboflow platform.',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'local', '', '1615FineTunedCapricciosaM_0812.pt'),
       ('1598FineTuned16150312_0101', '1598FineTuned16150312_0101',
        '1615FineTunedCapricciosaM_0312 fine tuned on 20 pages from Paris 1598a. The dataset was de-skewed before 1615FineTunedCapricciosaM_0312 was applied to it. After applying the segmentation, I also did some post processing: \n(1) Remove categories: "MainZone-P--Italics", "MainZone-P--Enunciation", "MainZone-P"\n(2) Remove overlap with 1000 precision: "DigitizationArtefactZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "MainZone", "MainZone-Head--Book", "MainZone-Head--Section", "NumberingZone", "QuireMarksZone", "RunningTitleZone"\nWhen I annotated in Roboflow, I further added a new category, that was not in the original model: "DropCapitalZone-Plane". This category was used for drop capitals that were not decorated. \nThe ground truth annotations ID ht01bz',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'local', '', '1598FineTuned16150312_0101.pt'),
       ('Gallicorpor', 'Gallicorpor', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'text', 'local', '',
        'Gallicorpor.mlmodel'),
       ('1615FineTunedGallicorpor_0301', '1615FineTunedGallicorpor_0301',
        'Gallicorpor fine tuned on 10 pages from Paris 1615. Annotations in...', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
        'text', 'local', '', '1615FineTunedGallicorpor_0301.mlmodel'),
       ('Paris1615NoContinuedPNoMainZone3', 'paris-1615-nocontinuedpnomainzone-dbxgq/3', '', CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP, 'segment', 'roboflow', 'yolo', ''),
       ('Paris1615PolygonsAndMainZone', 'paris-1615-polygonswithmz-wsrge/1', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
        'segment', 'roboflow', 'yolo', ''),
       ('Paris1615NoMainZoneSubtypes', 'paris-1615-withmznosubtypes-tkgii/1', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
        'segment', 'roboflow', 'yolo', ''),
       ('0212-xcfg-2', '0212-xcfg/2', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'roboflow', 'yolo', ''),
       ('1570-english-2', '1570-english/2', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'roboflow', 'yolo',
        ''),
       ('segmontoRB', 'segmonto/31', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'segment', 'roboflow', 'yolo', '');

-- model_categories (only segmontoRB has explicit categories)
INSERT INTO model_categories (model_id, category)
VALUES ('segmontoRB', 'AdvertisementZone'),
       ('segmontoRB', 'DigitizationArtefactZone'),
       ('segmontoRB', 'DropCapitalZone'),
       ('segmontoRB', 'FigureZone'),
       ('segmontoRB', 'FigureZone-FigDesc'),
       ('segmontoRB', 'FigureZone-Head'),
       ('segmontoRB', 'FormZone'),
       ('segmontoRB', 'GraphicZone'),
       ('segmontoRB', 'GraphicZone-Decoration'),
       ('segmontoRB', 'GraphicZone-FigDesc'),
       ('segmontoRB', 'GraphicZone-Head'),
       ('segmontoRB', 'GraphicZone-Maths'),
       ('segmontoRB', 'GraphicZone-Part'),
       ('segmontoRB', 'GraphicZone-TextualContent'),
       ('segmontoRB', 'MainZone-Continued'),
       ('segmontoRB', 'MainZone-Date'),
       ('segmontoRB', 'MainZone-Entry'),
       ('segmontoRB', 'MainZone-Entry-Continued'),
       ('segmontoRB', 'MainZone-Head'),
       ('segmontoRB', 'MainZone-Lg'),
       ('segmontoRB', 'MainZone-Lg-Continued'),
       ('segmontoRB', 'MainZone-List-Continued'),
       ('segmontoRB', 'MainZone-ListItem'),
       ('segmontoRB', 'MainZone-Maths'),
       ('segmontoRB', 'MainZone-Other'),
       ('segmontoRB', 'MainZone-P'),
       ('segmontoRB', 'MainZone-P-Continued'),
       ('segmontoRB', 'MainZone-Signature'),
       ('segmontoRB', 'MainZone-Sp'),
       ('segmontoRB', 'MainZone-Sp-Continued'),
       ('segmontoRB', 'MarginTextZone-ContinuedNotes'),
       ('segmontoRB', 'MarginTextZone-ManuscriptAddendum'),
       ('segmontoRB', 'MarginTextZone-Notes'),
       ('segmontoRB', 'MarginTextZone-Notes-Continued'),
       ('segmontoRB', 'MusicZone'),
       ('segmontoRB', 'NumberingZone'),
       ('segmontoRB', 'PageTitleZone'),
       ('segmontoRB', 'PageTitleZone-Index'),
       ('segmontoRB', 'QuireMarksZone'),
       ('segmontoRB', 'RunningTitleZone'),
       ('segmontoRB', 'StampZone'),
       ('segmontoRB', 'StampZone-Sticker'),
       ('segmontoRB', 'TableZone'),
       ('segmontoRB', 'TableZone-Continued'),
       ('segmontoRB', 'TableZone-Head'),
       ('segmontoRB', 'TitlePageZone'),
       ('segmontoRB', 'TitlePageZone-Index');

COMMIT;
