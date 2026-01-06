CREATE TABLE IF NOT EXISTS editions (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS facsimiles (
    edition_id TEXT NOT NULL,
    id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    url TEXT NOT NULL,
    PRIMARY KEY (edition_id, id),
    FOREIGN KEY (edition_id) REFERENCES editions(id) ON DELETE CASCADE
);

-- editions
INSERT INTO editions (id, name, description, created_at, updated_at)
VALUES
    ('Paris_1615', 'Paris_1615', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Paris_1598a', 'Paris_1598a', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('London_1570', 'London_1570', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- facsimiles: Paris_1615
INSERT INTO facsimiles (edition_id, id, name, description, url, created_at, updated_at)
VALUES
    (
        'Paris_1615',
        '1',
        'Paris_1615 facsimile 1',
        '',
        'https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1615.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'Paris_1615',
        '2',
        'Paris_1615 facsimile 2',
        '',
        'https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1615.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

-- facsimiles: Paris_1598a
INSERT INTO facsimiles (edition_id, id, name, description, url, created_at, updated_at)
VALUES
    (
        'Paris_1598a',
        '1',
        'Paris_1598a facsimile 1',
        '',
        'https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1598a.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'Paris_1598a',
        '2',
        'Paris_1598a facsimile 2',
        '',
        'https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/Paris_1598a.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

-- facsimiles: London_1570
INSERT INTO facsimiles (edition_id, id, name, description, url, created_at, updated_at)
VALUES
    (
        'London_1570',
        '1',
        'London_1570 facsimile 1',
        '',
        'https://github.com/ReallyLiri/elements-facsimile/blob/main/docs/London_1570.pdf',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );
