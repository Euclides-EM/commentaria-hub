package store

import (
	"database/sql"
)

type BaseSQL struct {
	DB *sql.DB
}
