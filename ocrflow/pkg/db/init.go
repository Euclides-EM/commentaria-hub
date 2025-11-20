package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

//go:embed schema_mig.sql
var schemaSQL string

func InitDB(dbPath, migrationsDir string) (*sql.DB, error) {
	if err := initDBFile(dbPath); err != nil {
		return nil, fmt.Errorf("failed to init DB file: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	if err := migrateDB(db, migrationsDir); err != nil {
		defer db.Close()
		return nil, fmt.Errorf("failed to migrate DB: %w", err)
	}

	return db, nil
}

func initDBFile(dbPath string) error {
	if dbPath == "" {
		return fmt.Errorf("database path is empty: %s", dbPath)
	}

	_, err := os.Stat(dbPath)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	f, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

func migrateDB(db *sql.DB, migrationsDir string) error {
	log.Printf("initializing schema migrations table...")
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to init schema migrations table: %w", err)
	}

	migrations, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	slices.SortFunc(migrations, func(a, b os.DirEntry) int {
		ats, err := migrationFileNameToTimestamp(a.Name())
		if err != nil {
			log.Fatalf("invalid migration file name: %s", a.Name())
		}
		bts, err := migrationFileNameToTimestamp(b.Name())
		if err != nil {
			log.Fatalf("invalid migration file name: %s", b.Name())
		}
		return ats - bts
	})

	appliedMigrations := 0
	for _, migration := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(1) FROM schema_migrations WHERE name = ?", migration.Name()).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check whether migration %s has been applied: %w", migration.Name(), err)
		}
		if count > 0 {
			continue
		}

		log.Printf("applying migration %s", migration.Name())
		migrationPath := path.Join(migrationsDir, migration.Name())
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
		}

		if _, err := db.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationPath, err)
		}

		if _, err = db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migration.Name()); err != nil {
			return fmt.Errorf("failed to record applied migration %s: %w", migration.Name(), err)
		}

		appliedMigrations += 1
	}

	log.Printf("%d new DB migrations applied, db is up to date", appliedMigrations)
	return nil
}

func migrationFileNameToTimestamp(fileName string) (int, error) {
	parts := strings.Split(fileName, "_")
	return strconv.Atoi(parts[0])
}
