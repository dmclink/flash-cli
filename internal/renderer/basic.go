package renderer

import (
	"context"
	"fmt"

	"github.com/dmclink/flash-cli/internal/database"
)

type BasicRenderer struct{}

func (r BasicRenderer) Render(ctx context.Context, card database.Flashcard, cardNum int, cardCount int, unparsedMods []string) (string, string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", "", err
	}

	return fmt.Sprintf("Front:\n%s", card.Front),
		fmt.Sprintf("Front:\n%s\n\nBack:\n%s\n\n", card.Front, card.Back),
		"",
		nil
}
