package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/viper"
)

var V = viper.New()

const (
	KeyAddSeparator          = "add.separator"
	KeyDefaultFilterGroup    = "default.filter.groups"
	KeyDefaultFilterTag      = "default.filter.tags"
	KeyDefaultReviewMode     = "default.review.mode"
	KeyDefaultReviewRenderer = "default.review.renderer"
	KeyDefaultReviewLimit    = "default.review.limit"
	KeyPathPlugins           = "path.plugins"
	KeyPathLogs              = "path.logs"
)

func InitConfig() error {
	err := setInitDefaults()
	if err != nil {
		return fmt.Errorf("setting config defaults | %w", err)
	}

	var configPath string

	if envPath := os.Getenv("FLASH_CLI_CONFIG_DIR"); envPath != "" {
		configPath = envPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to find home directory for os | %w", err)
		}
		configPath = filepath.Join(home, ".config", constant.APP_NAME)
	}

	V.SetConfigName("config")
	V.SetConfigType("yaml")
	V.AddConfigPath(configPath)
	V.AutomaticEnv()

	err = V.ReadInConfig()
	if err == nil {
		return nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		if err := os.MkdirAll(configPath, 0o755); err != nil {
			return fmt.Errorf("creating config directory | %w", err)
		}

		if err := V.SafeWriteConfig(); err != nil {
			return fmt.Errorf("creating default config | %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed reading config file | %w", err)
}

func setInitDefaults() error {
	V.SetDefault(KeyAddSeparator, "::")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to find home directory for os | %w", err)
	}
	defaultConfigDir := filepath.Join(home, ".config", constant.APP_NAME)
	V.SetDefault(KeyPathPlugins, filepath.Join(defaultConfigDir, "plugins"))
	V.SetDefault(KeyPathLogs, filepath.Join(home, ".local", "state", constant.APP_NAME, "plugins.log"))

	V.SetDefault(KeyDefaultFilterGroup, "")
	V.SetDefault(KeyDefaultFilterTag, "")
	V.SetDefault(KeyDefaultReviewMode, "")
	V.SetDefault(KeyDefaultReviewRenderer, "")
	V.SetDefault(KeyDefaultReviewLimit, "")
	return nil
}

func SetUserDefault(key string, value interface{}) error {
	V.Set(key, value)
	if err := V.WriteConfig(); err != nil {
		return fmt.Errorf("persisting new config value | %w", err)
	}

	return nil
}
