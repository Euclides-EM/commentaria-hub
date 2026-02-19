package db

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema_mig.sql
var schemaSQL string

func InitDB(dbPath string, migrations embed.FS, migrationsSubdir string) (*sql.DB, error) {
	if err := initDBFile(dbPath); err != nil {
		return nil, fmt.Errorf("failed to init DB file: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	// Verify connection early
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Sensible defaults
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err = migrateDB(db, migrations, migrationsSubdir); err != nil {
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

func migrateDB(db *sql.DB, migrationsFS fs.FS, migrationsSubdir string) error {
	log.Printf("initializing schema migrations table...")
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to init schema migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, migrationsSubdir)
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	// Filter to .sql files only and sort by timestamp prefix
	var migrations []fs.DirEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		migrations = append(migrations, e)
	}
	slices.SortFunc(migrations, func(a, b fs.DirEntry) int {
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
		migrationPath := migrationsSubdir + "/" + migration.Name()
		migrationSQL, err := fs.ReadFile(migrationsFS, migrationPath)
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
