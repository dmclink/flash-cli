package database

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
)

var schema = fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %[1]s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BLOB UNIQUE,
		last_review INTEGER DEFAULT (unixepoch()),
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		created_at INTEGER DEFAULT (unixepoch()),
		ext_data BLOB
	);

	CREATE TABLE IF NOT EXISTS %[2]s (
		flashcard_id INTEGER NOT NULL,
		tag TEXT NOT NULL,
		PRIMARY KEY (tag, flashcard_id),
		FOREIGN KEY REFERENCES %[1]s(id) ON DELETE CASCADE
	) WITHOUT ROWID;

	CREATE TABLE IF NOT EXISTS %[3]s (
		flashcard_id INTEGER NOT NULL,
		group_name TEXT NOT NULL,
		PRIMARY KEY (group_name, flashcard_id),
		FOREIGN KEY REFERENCES %[1]s(id) ON DELETE CASCADE
	) WITHOUT ROWID;

	-- secondary index for pulling data on edit
	CREATE INDEX IF NOT EXISTS idx_%[2]s_card_id ON %[2]s(flashcard_id);
	CREATE INDEX IF NOT EXISTS idx_%[3]s_card_id ON %[3]s(flashcard_id);
	`, constant.DATABASE_TABLE_FLASHCARDS, constant.DATABASE_TABLE_TAGS, constant.DATABASE_TABLE_GROUPS)
