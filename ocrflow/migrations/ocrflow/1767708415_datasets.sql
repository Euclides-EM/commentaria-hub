CREATE TABLE IF NOT EXISTS datasets (
    id TEXT PRIMARY KEY NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    edition_id TEXT NOT NULL,
    facsimile_id TEXT NOT NULL,
    dpi INTEGER NOT NULL,
    deskewed BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (edition_id, facsimile_id) REFERENCES facsimiles(edition_id, id) ON DELETE CASCADE
);

-- datasets
INSERT INTO datasets (
    id,
    name,
    description,
    created_at,
    updated_at,
    edition_id,
    facsimile_id,
    dpi,
    deskewed
) VALUES
      (
          'rrpbnk',
          'Paris 1615 No Alignment',
          '',
          CURRENT_TIMESTAMP,
          CURRENT_TIMESTAMP,
          'Paris_1615',
          '2',
          300,
          FALSE
      ),
      (
          'uk5wbj',
          'Paris 1615 Deskewed',
          '',
          CURRENT_TIMESTAMP,
          CURRENT_TIMESTAMP,
          'Paris_1615',
          '1',
          300,
          TRUE
      ),
      (
          'aiqcec',
          'London 1570 Deskewed',
          '',
          CURRENT_TIMESTAMP,
          CURRENT_TIMESTAMP,
          'London_1570',
          '1',
          300,
          TRUE
      ),
      (
          'nu3e82',
          'London 1570 No Alignment',
          '',
          CURRENT_TIMESTAMP,
          CURRENT_TIMESTAMP,
          'Paris_1598a',
          '1',
          300,
          FALSE
      ),
      (
          'mq9w7q',
          'Paris 1598a Deskewed',
          '',
          CURRENT_TIMESTAMP,
          CURRENT_TIMESTAMP,
          'Paris_1598a',
          '2',
          300,
          TRUE
      );

