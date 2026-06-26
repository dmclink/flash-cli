package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dmclink/flash-cli/internal/args"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

var modsStartingIdxKey = "modsStartingIdx"

func Execute() error {
	reorderedArgs, idx, err := args.Reorder(os.Args)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: reordering args | %w", err))
		// TODO: call usage?
		os.Exit(1)
	}
	os.Args = reorderedArgs

	ctx := context.WithValue(context.Background(), modsStartingIdxKey, idx)

	return rootCmd.ExecuteContext(ctx)
}

var (
	cfgFile string
	DB      *sql.DB

	rootCmd = &cobra.Command{
		Use:   constant.APP_NAME,
		Short: "Flashcard review and management program",
		Long:  "A CLI program to review and manage flashcards backed by an SQLite database. Strives for simplicity and ease of use to add and review. Extensible via plugins.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if os.Getuid() == 0 {
				fmt.Println("Error: do not run this application as root/sudo.")
				os.Exit(1)
			}

			var err error
			DB, err = database.OpenAndInitDatabase()
			if err != nil {
				fmt.Println("Error: Opening and initializing database")
				fmt.Println(err)
				os.Exit(1)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: run the default command when calling root by itself, likely reviewCmd
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if DB != nil {
				DB.Close()
			}
		},
		Version: "0.1.0",
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	// TODO: define additional global flags here
}

// TODO: this is boilerplate from the cobra docs, review later if we ever used config files and either delete this block or delete this comment
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
