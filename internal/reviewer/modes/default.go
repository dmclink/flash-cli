package reviewer

import (
	"context"

	"github.com/dmclink/flash-cli/internal/database"
)

type LinearMode struct{}

// doesn't do anything just spits the cards out as they came in
func (m LinearMode) Process(ctx context.Context, cards []database.Flashcard, modifiers []string) ([]database.Flashcard, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}
