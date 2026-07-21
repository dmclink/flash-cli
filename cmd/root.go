package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmclink/flash-cli/internal/config"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/logger"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

func Execute(db *sql.DB) error {
	rootCmd := NewRootCmd(db)

	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	parsedArgs, err := parser.ParseArgs(os.Args)
	if err != nil {
		// TODO: call usage for parsed command? would need some map or something to lookup command funcs
		rootCmd.Usage()
		return fmt.Errorf("failed to validate and reorder args | %w", err)
	}
	os.Args = parsedArgs.Args(os.Args[0])

	// TODO: consider maintaining a global (or context) commands set that gets built here
	// to use for FindCommand in the parser instead of current naive implementation to
	// stop at the first word that doesn't match a filter

	rootCmd.AddCommand(NewVersionCmd(db))
	rootCmd.AddCommand(NewAddCmd(db))
	rootCmd.AddCommand(NewReviewCmd(db))

	ctx, stop := signal.NotifyContext(context.WithValue(context.Background(), constant.PARSED_ARGS_KEY, parsedArgs), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// TODO: remove error from signature and just call os.Exit(1) instead?
	// compare the different default output behavior from cobra on erroring with both
	return rootCmd.ExecuteContext(ctx)
}

func NewRootCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:                constant.APP_NAME,
		Short:              "Flashcard review and management program",
		Long:               "A CLI program to review and manage flashcards backed by an SQLite database. Strives for simplicity and ease of use to add and review. Extensible via plugins.",
		DisableFlagParsing: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := config.InitConfig()
			if err != nil {
				return fmt.Errorf("initializing viper config | %w", err)
			}

			if err := logger.InitPluginLogger(); err != nil {
				return fmt.Errorf("initializing plugin logger | %w", err)
			}

			return nil
		},
		Version: "0.1.0",
	}
}
