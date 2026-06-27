package database

const schema = `
	CREATE TABLE IF NOT EXISTS flashcards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid BLOB UNIQUE,
		last_reviewed INTEGER DEFAULT (unixepoch()),
		front TEXT NOT NULL,
		back TEXT NOT NULL,
		created_at INTEGER DEFAULT (unixepoch()),
		ext_data BLOB
	);`
