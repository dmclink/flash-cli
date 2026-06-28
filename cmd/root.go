package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

func Execute(db *sql.DB) error {
	v := viper.New()

	var cfgFile string

	rootCmd := NewRootCmd(db, v, &cfgFile)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	parsedArgs, err := parser.ParseArgs(os.Args)
	if err != nil {
		// TODO: call usage for parsed command? would need some map or something to lookup command funcs
		rootCmd.Usage()
		return fmt.Errorf("failed to validate and reorder args | %w", err)
	}
	os.Args = parsedArgs.Args

	rootCmd.AddCommand(NewVersionCmd(db, v))
	rootCmd.AddCommand(NewAddCmd(db, v))

	ctx := context.WithValue(context.Background(), constant.PARSED_ARGS_KEY, parsedArgs)

	// TODO: remove error from signature and just call os.Exit(1) instead?
	return rootCmd.ExecuteContext(ctx)
}

func NewRootCmd(db *sql.DB, v *viper.Viper, cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   constant.APP_NAME,
		Short: "Flashcard review and management program",
		Long:  "A CLI program to review and manage flashcards backed by an SQLite database. Strives for simplicity and ease of use to add and review. Extensible via plugins.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := initConfig(v, cfgFile)
			if err != nil {
				return fmt.Errorf("initializing viper config | %w", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: run the default command when calling root by itself, likely reviewCmd
		},
		Version: "0.1.0",
	}
}

// func init() {
// 	cobra.OnInitialize(initConfig)
//
//
// 	// TODO: define additional global flags here
// }

// TODO: this is boilerplate from the cobra docs, review later if we ever used config files and either delete this block or delete this comment
func initConfig(v *viper.Viper, cfgFile *string) error {
	v.SetDefault(constant.VIPER_KEY_DELIMITER, "::")

	var configPath string
	configName := "config"
	configType := "yaml"
	if *cfgFile != "" {
		configPath = filepath.Dir(*cfgFile)
		configName = strings.TrimSuffix(filepath.Base(*cfgFile), filepath.Ext(*cfgFile))
		configType = strings.TrimPrefix(filepath.Ext(*cfgFile), ".")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory for os | %w", err)
		}

		configPath = filepath.Join(home, ".config", constant.APP_NAME)
	}

	v.SetConfigName(configName)
	v.SetConfigType(configType)
	v.AddConfigPath(configPath)
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err == nil {
		return nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		if err := os.MkdirAll(configPath, 0o755); err != nil {
			return fmt.Errorf("failed creating config directory | %w", err)
		}

		if err := v.SafeWriteConfig(); err != nil {
			return fmt.Errorf("failed creating default config | %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed reading config file | %w", err)
}
