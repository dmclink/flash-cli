package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/google/uuid"
)

// TODO: add doc comments
// TODO: add  plugin ext_data to args and to database exec
func AddFlashcard(db *sql.DB, front, back string, groups, tags []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	flashcardsQuery := fmt.Sprintf(`INSERT INTO %s (uuid, front, back, ext_data) VALUES (?, ?, ?, ?)`, constant.DATABASE_TABLE_FLASHCARDS)
	uuid := uuid.New().String()
	// TODO: don't just pass an empty {} here if plugin data is included
	res, err := tx.Exec(flashcardsQuery, uuid, front, back, `{}`)
	if err != nil {
		return fmt.Errorf("adding card to db | %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("pulling last insert id | %w", err)
	}

	if len(groups) > 0 {
		for _, group := range groups {
			groupsQuery := fmt.Sprintf(`INSERT INTO %s (flashcard_id, group_name) VALUES (?, ?)`, constant.DATABASE_TABLE_GROUPS)
			_, err = tx.Exec(groupsQuery, id, group)
			if err != nil {
				return fmt.Errorf("adding groups to db | %w", err)
			}
		}
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			tagsQuery := fmt.Sprintf("INSERT INTO %s (flashcard_id, tag) VALUES (?, ?)", constant.DATABASE_TABLE_TAGS)
			_, err = tx.Exec(tagsQuery, id, tag)
			if err != nil {
				return fmt.Errorf("adding tags to db | %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing database transaction | %w", err)
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
