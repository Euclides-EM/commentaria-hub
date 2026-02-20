create table datasets
(
    id           TEXT                  not null primary key,
    name         VARCHAR(255)          not null,
    description  TEXT    default ''    not null,
    created_at   TIMESTAMP             not null,
    updated_at   TIMESTAMP             not null,
    edition_id   TEXT                  not null,
    facsimile_id TEXT                  not null,
    dpi          INTEGER               not null,
    deskewed     BOOLEAN default FALSE not null,
    foreign key (edition_id, facsimile_id) references facsimiles on delete cascade
);
