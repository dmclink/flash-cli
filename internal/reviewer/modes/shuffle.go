package reviewer

import (
	"context"
	"math/rand/v2"

	"github.com/dmclink/flash-cli/internal/database"
)

type ShuffleMode struct{}

func (m ShuffleMode) Process(ctx context.Context, cardsIn []database.Flashcard, modifiers []string) ([]database.Flashcard, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cardsOut := make([]database.Flashcard, len(cardsIn))
	copy(cardsOut, cardsIn)

	rand.Shuffle(len(cardsOut), func(i, j int) { cardsOut[i], cardsOut[j] = cardsOut[j], cardsOut[i] })

	return cardsOut, nil
}
