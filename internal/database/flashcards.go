package database

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/google/uuid"
)

func AddFlashcard(db *sql.DB, front, back string) error {
	uuid := uuid.New().String()
	// TODO: need to add external data after that is implemented
	query := fmt.Sprintf(`INSERT INTO %s (uuid, front, back) VALUES (?, ?, ?)`, constant.DATABASE_FLASHCARDS_TABLE)

	_, err := db.Exec(query, uuid, front, back)
	if err != nil {
		return fmt.Errorf("adding card to db | %w", err)
	}

	return nil
}

func GetAllFlashcards(db *sql.DB) ([]Flashcard, error) {
	query := fmt.Sprintf(`
		SELECT id, uuid, last_review, front, back, created_at, ext_data FROM %s;
	`, constant.DATABASE_FLASHCARDS_TABLE)

	rows, err := db.Query(query)
	if err != nil {
		return []Flashcard{}, fmt.Errorf("failed performing database query | %w", err)
	}

	result := []Flashcard{}
	for rows.Next() {
		fc := Flashcard{}
		err := rows.Scan(&fc.ID, &fc.UUID, &fc.LastReview, &fc.Front, &fc.Back, &fc.CreatedAt, &fc.ExtData)
		if err != nil {
			return []Flashcard{}, fmt.Errorf("failed scanning row | %w", err)
		}

		result = append(result, fc)
	}
	return result, nil
}
