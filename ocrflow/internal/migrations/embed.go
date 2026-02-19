package migrations

import "embed"

// Migrations is the embedded filesystem of SQL migration files (ocrflow/*.sql).
// Used by pkg/db to run migrations at startup without a filesystem path.
//
//go:embed ocrflow/*.sql
var Migrations embed.FS
