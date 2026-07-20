package reviewer

import (
	"context"
	"sort"

	"github.com/dmclink/flash-cli/internal/database"
)

type CreatedAtMode struct{}

func (m CreatedAtMode) Process(ctx context.Context, cardsIn []database.Flashcard, modifiers []string) ([]database.Flashcard, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cardsOut := make([]database.Flashcard, len(cardsIn))
	copy(cardsOut, cardsIn)

	sort.Slice(cardsOut, func(i, j int) bool {
		return cardsOut[i].CreatedAt < cardsOut[j].CreatedAt
	})

	return cardsOut, nil
}
