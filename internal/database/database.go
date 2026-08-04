package database

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/dmclink/flash-cli/internal/platform"
)

// Open opens sqlite database at path appropriate for user's operating system
func Open() (*sql.DB, error) {
	path := DatabasePath()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Init executes creation of any tables for the database if they do not exist
func Init(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		db.Close()
		return err
	}

	return nil
}

func DatabasePath() string {
	dataDir := platform.DataDir()

	return filepath.Join(dataDir, "app.db")
}
