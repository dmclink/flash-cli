package cmd

import (
	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func NewRootCmd(a *app.App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                constant.APP_NAME,
		Short:              "Flashcard review and management program",
		Long:               "A CLI program to review and manage flashcards backed by an SQLite database. Strives for simplicity and ease of use to add and review. Extensible via plugins.",
		DisableFlagParsing: true,
		Version:            constant.VERSION,
	}

	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	// TODO: consider maintaining a global (or context) commands set that gets built here
	// to use for FindCommand in the parser instead of current naive implementation to
	// stop at the first word that doesn't match a filter

	rootCmd.AddCommand(NewVersionCmd(a))
	rootCmd.AddCommand(NewAddCmd(a))
	rootCmd.AddCommand(NewReviewCmd(a))
	rootCmd.AddCommand(NewConfigCmd(a))

	return rootCmd
}
