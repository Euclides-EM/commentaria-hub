package store

import (
	"database/sql"
)

type BaseSQL struct {
	db *sql.DB
}
