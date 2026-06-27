package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/platform"
)

// Open opens sqlite database at path appropriate for user's operating system
func Open() (*sql.DB, error) {
	path, err := DatabasePath()
	if err != nil {
		return nil, err
	}

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
