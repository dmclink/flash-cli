package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

func Execute() error {
	parsedArgs, err := parser.ParseArgs(os.Args)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: reordering args | %w", err))
		// TODO: call usage?, if not calling usage, can optionally handle error in main
		// TODO: also look up if we should handle errors in main for cobra, since the bottom of this function
		// returns an error, should it be handled? Cobra automatically pushes to stderr anyway do i need to do anything with this?
		os.Exit(1)
	}
	os.Args = parsedArgs.Args

	ctx := context.WithValue(context.Background(), constant.PARSED_ARGS_KEY, parsedArgs)

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
				return fmt.Errorf("Error: do not run this application as root/sudo.")
			}

			var err error
			DB, err = database.OpenAndInitDatabase()
			if err != nil {
				return fmt.Errorf("Error: Opening and initializing database | %w", err)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: run the default command when calling root by itself, likely reviewCmd
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if DB != nil {
				return DB.Close()
			}
			return fmt.Errorf("nil database")
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
