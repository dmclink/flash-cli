package reviewer

import (
	"context"
	"sort"

	"github.com/dmclink/flash-cli/internal/database"
)

type LastReviewMode struct{}

func (m LastReviewMode) Process(ctx context.Context, cardsIn []database.Flashcard, modifiers []string) ([]database.Flashcard, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cardsOut := make([]database.Flashcard, len(cardsIn))
	copy(cardsOut, cardsIn)

	sort.Slice(cardsOut, func(i, j int) bool {
		return cardsOut[i].LastReview < cardsOut[j].LastReview
	})

	return cardsOut, nil
}
