package cmd

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/reviewer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewReviewCmd(db *sql.DB, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "review [group|id filters]... [reviewing style mods]...",
		Short: "Review flashcards",
		// TODO: change these comments about mods, filters, and config after those are implemented
		Long: "Review flashcards in order by set by mods or defaults ordered by last reviewed, oldest first. Shows one flashcard at a time. Can be filtered by groups or ID ranges. Settings can be changed with config",
		// TODO: parse filters and mods here
		// PreRunE: func(cmd *cobra.Command, args []string) error {
		// 	return nil
		// },
		RunE: func(cmd *cobra.Command, args []string) error {
			cards, err := database.GetAllFlashcards(db)
			if err != nil {
				return fmt.Errorf("getting flashcards from db | %w", err)
			}

			err = reviewer.Review(cards)
			if err != nil {
				return fmt.Errorf("reviewing cards | %w", err)
			}
			// for _, card := range cards {
			// 	fmt.Println(card)
			// }
			return nil
		},
	}
}
