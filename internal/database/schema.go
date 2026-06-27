package database

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
)

var schema = fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BLOB UNIQUE,
		last_review INTEGER DEFAULT (unixepoch()),
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		created_at INTEGER DEFAULT (unixepoch()),
		ext_data BLOB
	);`, constant.DATABASE_FLASHCARDS_TABLE)
