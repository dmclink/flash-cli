package database

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
)

var schema = fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %[1]s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BLOB UNIQUE,
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		last_review INTEGER DEFAULT (unixepoch()),
		created_at INTEGER DEFAULT (unixepoch()),
		ext_data TEXT CHECK (json_valid(ext_data))
	);

	CREATE TABLE IF NOT EXISTS %[2]s (
		flashcard_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		PRIMARY KEY (name, flashcard_id),
		FOREIGN KEY (flashcard_id) REFERENCES %[1]s(id) ON DELETE CASCADE
	) WITHOUT ROWID;

	CREATE TABLE IF NOT EXISTS %[3]s (
		flashcard_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		PRIMARY KEY (name, flashcard_id),
		FOREIGN KEY (flashcard_id) REFERENCES %[1]s(id) ON DELETE CASCADE
	) WITHOUT ROWID;

	-- secondary index for pulling data on edit
	CREATE INDEX IF NOT EXISTS idx_%[2]s_card_id ON %[2]s(flashcard_id);
	CREATE INDEX IF NOT EXISTS idx_%[3]s_card_id ON %[3]s(flashcard_id);
	`, constant.DATABASE_TABLE_FLASHCARDS, constant.DATABASE_TABLE_TAGS, constant.DATABASE_TABLE_GROUPS)

// TODO: consider creating a plugin registry metadata table for plugin names and field keys for autocomplete
// and ensuring custom field types are valid
/* ie:
CREATE TABLE IF NOT EXISTS %s_plugin_registry (
	plugin_name TEXT NOT NULL,
	field_key TEXT NOT NULL,
	PRIMARY KEY (plugin_name, field_key)
) WITHOUT ROWID;
*/
