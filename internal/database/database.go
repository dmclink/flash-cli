package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/platform"
)

func OpenAndInitDatabase() (*sql.DB, error) {
	path, err := DatabasePath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS flashcards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BLOB UNIQUE,
		last_reviewed INTEGER DEFAULT (unixepoch()),
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		created_at INTEGER DEFAULT (unixepoch()),
		ext_data BLOB
	);`

	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func DatabasePath() (string, error) {
	dataDir, err := platform.DataDirectory()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(dataDir, constant.APP_NAME)
	os.UserConfigDir()

	err = os.MkdirAll(appDir, 0o755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dataDir, constant.APP_NAME, "app.db"), nil
}
