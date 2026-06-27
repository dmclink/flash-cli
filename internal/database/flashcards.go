package database

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/google/uuid"
)

func AddFlashcard(db *sql.DB, front, back string) error {
	uuid := uuid.New().String()
	query := fmt.Sprintf(`INSERT INTO %s (uuid, front, back) VALUES (?, ?, ?)`, constant.DATABASE_FLASHCARDS_TABLE)

	_, err := db.Exec(query, uuid, front, back)
	if err != nil {
		return fmt.Errorf("adding card to db | %w", err)
	}

	return nil
}
