package database

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/google/uuid"
)

// TODO: add comments and add custom
func AddFlashcard(db *sql.DB, front, back string, groups, tags []string) error {
	uuid := uuid.New().String()
	query := fmt.Sprintf(`INSERT INTO %s (uuid, front, back, ext_data) VALUES (?, ?, ?, ?)`, constant.DATABASE_TABLE_FLASHCARDS)

	// TODO: add groups and tags and plugin data to args

	// TODO: don't just pass an empty {} here if plugin data is included
	_, err := db.Exec(query, uuid, front, back, `{}`)
	if err != nil {
		return fmt.Errorf("adding card to db | %w", err)
	}

	return nil
}

func GetAllFlashcards(db *sql.DB) ([]Flashcard, error) {
	query := fmt.Sprintf(`
		SELECT id, uuid, last_review, front, back, created_at, ext_data FROM %s;
	`, constant.DATABASE_TABLE_FLASHCARDS)

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
