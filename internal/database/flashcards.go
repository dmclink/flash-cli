package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/google/uuid"
)

func UpdateFlashcard(ctx context.Context, db *sql.DB, oldCard FullFlashcard, newCard FullFlashcard) error {
	if oldCard.ID != newCard.ID {
		return fmt.Errorf("cannot update ID or different card passed: old:%d new:%d", oldCard.ID, newCard.ID)
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`UPDATE %s SET front = ?, back = ?, last_review = ?, created_at = ?, ext_data = json(?) WHERE id = ?`, constant.DATABASE_TABLE_FLASHCARDS)
	args := []any{newCard.Front, newCard.Back, newCard.LastReview, newCard.CreatedAt, newCard.ExtData, newCard.ID}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(oldCard.Groups, newCard.Groups) {
		query := fmt.Sprintf(`DELETE FROM %s WHERE flashcard_id = ?`, constant.DATABASE_TABLE_GROUPS)
		args := []any{newCard.ID}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		if len(newCard.Groups) > 0 {
			for _, group := range newCard.Groups {
				groupsQuery := fmt.Sprintf(`INSERT INTO %s (flashcard_id, name) VALUES (?, ?)`, constant.DATABASE_TABLE_GROUPS)
				_, err = tx.Exec(groupsQuery, newCard.ID, group)
				if err != nil {
					return fmt.Errorf("adding groups to db | %w", err)
				}
			}
		}

	}

	if !reflect.DeepEqual(oldCard.Tags, newCard.Tags) {
		query := fmt.Sprintf(`DELETE FROM %s WHERE flashcard_id = ?`, constant.DATABASE_TABLE_TAGS)
		args := []any{newCard.ID}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		if len(newCard.Tags) > 0 {
			for _, tag := range newCard.Tags {
				tagsQuery := fmt.Sprintf("INSERT INTO %s (flashcard_id, name) VALUES (?, ?)", constant.DATABASE_TABLE_TAGS)
				_, err = tx.Exec(tagsQuery, newCard.ID, tag)
				if err != nil {
					return fmt.Errorf("adding tags to db | %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// TODO: add doc comments
// TODO: add  plugin ext_data to args and to database exec
func AddFlashcard(db *sql.DB, front, back string, groups, tags []parser.Filter) (int, error) {
	// TODO: this should be using the cancellable cobra context from main.go instead of creating a separate one
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	flashcardsQuery := fmt.Sprintf(`INSERT INTO %s (uuid, front, back, ext_data) VALUES (?, ?, ?, ?)`, constant.DATABASE_TABLE_FLASHCARDS)
	uuid := uuid.New().String()
	// TODO: don't just pass an empty {} here if plugin data is included
	res, err := tx.Exec(flashcardsQuery, uuid, front, back, `{}`)
	if err != nil {
		return -1, fmt.Errorf("adding card to db | %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("pulling last insert id | %w", err)
	}

	if len(groups) > 0 {
		for _, group := range groups {
			groupsQuery := fmt.Sprintf(`INSERT INTO %s (flashcard_id, name) VALUES (?, ?)`, constant.DATABASE_TABLE_GROUPS)
			_, err = tx.Exec(groupsQuery, id, group.Value)
			if err != nil {
				return -1, fmt.Errorf("adding groups to db | %w", err)
			}
		}
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			if tag.IsExclude {
				fmt.Printf("ignoring excluded tag -%s. Did you mean to use a '+'?\n", tag.Value)
				continue
			}
			tagsQuery := fmt.Sprintf("INSERT INTO %s (flashcard_id, name) VALUES (?, ?)", constant.DATABASE_TABLE_TAGS)
			_, err = tx.Exec(tagsQuery, id, tag.Value)
			if err != nil {
				return -1, fmt.Errorf("adding tags to db | %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return -1, fmt.Errorf("committing database transaction | %w", err)
	}

	return int(id), nil
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

func GetFlashcards(db *sql.DB, filters parser.SearchFilters) ([]Flashcard, error) {
	query, args := buildFlashcardSelectQuery(filters)

	rows, err := db.Query(query, args...)
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

func getNames(db *sql.DB, id int, table string) ([]string, error) {
	zero := make([]string, 0)
	query := fmt.Sprintf("SELECT name FROM %s WHERE flashcard_id = ?", table)
	rows, err := db.Query(query, id)
	if err != nil {
		return zero, err
	}

	result := []string{}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return zero, err
		}
		result = append(result, name)
	}

	return result, nil
}

func GetFlashcardGroups(db *sql.DB, id int) ([]string, error) {
	return getNames(db, id, constant.DATABASE_TABLE_GROUPS)
}

func GetFlashcardTags(db *sql.DB, id int) ([]string, error) {
	return getNames(db, id, constant.DATABASE_TABLE_TAGS)
}

func buildFlashcardSelectQuery(filters parser.SearchFilters) (string, []any) {
	fTable := constant.DATABASE_TABLE_FLASHCARDS
	gTable := constant.DATABASE_TABLE_GROUPS
	tTable := constant.DATABASE_TABLE_TAGS

	// NOTE: casting fTable as f we will use this throughout our query strings
	baseQuery := fmt.Sprintf("SELECT\n\tf.id, f.uuid, f.last_review, f.front, f.back, f.created_at, json(f.ext_data)\nFROM\n\t%s f", fTable)

	size := filters.Size()
	if size == 0 {
		return baseQuery + ";", []any{}
	}

	andStrings := make([]string, 0, size)
	orStrings := make([]string, 0, size)
	queryArgs := []any{}

	for _, id := range filters.IDs {
		var buf strings.Builder
		buf.WriteString("f.id = ?")

		queryArgs = append(queryArgs, id.Low)
		orStrings = append(orStrings, buf.String())
	}

	for _, r := range filters.Ranges {
		var buf strings.Builder
		buf.WriteString("f.id BETWEEN ? AND ?")

		queryArgs = append(queryArgs, r.Low, r.High)
		orStrings = append(orStrings, buf.String())
	}

	for _, u := range filters.UUIDs {
		var buf strings.Builder
		buf.WriteString("f.uuid = ?")

		queryArgs = append(queryArgs, u.Value)
		orStrings = append(orStrings, buf.String())
	}

	for _, g := range filters.Groups {
		var buf strings.Builder
		buf.WriteString("EXISTS (SELECT 1 FROM " + gTable + " g WHERE f.id = g.flashcard_id AND g.name = ?)")

		queryArgs = append(queryArgs, g.Value)
		orStrings = append(orStrings, buf.String())
	}

	for _, t := range filters.Tags {
		var buf strings.Builder
		if t.IsExclude {
			buf.WriteString("NOT ")
		}
		buf.WriteString("EXISTS (SELECT 1 FROM " + tTable + " t WHERE f.id = t.flashcard_id AND t.name = ?)")

		queryArgs = append(queryArgs, t.Value)
		andStrings = append(andStrings, buf.String())
	}

	var whereQuery strings.Builder
	whereQuery.WriteString("WHERE\n")

	if len(orStrings) > 0 {
		whereQuery.WriteString("\t(\n")
		fmt.Fprintf(&whereQuery, "\t\t%s\n", orStrings[0])
		for i := 1; i < len(orStrings); i++ {
			fmt.Fprintf(&whereQuery, "\t\tOR %s\n", orStrings[i])
		}
		whereQuery.WriteString("\t)")
	}

	if len(andStrings) > 0 {
		if len(orStrings) > 0 {
			whereQuery.WriteString("\n\tAND (\n")
		} else {
			whereQuery.WriteString("\t(\n")
		}
		fmt.Fprintf(&whereQuery, "\t\t%s\n", andStrings[0])
		for i := 1; i < len(andStrings); i++ {
			fmt.Fprintf(&whereQuery, "\t\tAND %s\n", andStrings[i])
		}
		whereQuery.WriteString("\t)")
	}

	return fmt.Sprintf("%s\n%s;", baseQuery, whereQuery.String()), queryArgs
}
