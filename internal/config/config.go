package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	KeyPathPluginsDir        = "path.plugins_dir"
	KeyPathLogsDir           = "path.logs_dir"
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
		fmt.Println("configPath:", configPath)
	}

	V.SetConfigName("config")
	V.SetConfigType("yaml")
	V.AddConfigPath(configPath)
	V.SetEnvPrefix("FLASH_CLI")
	V.AutomaticEnv()

	err = V.ReadInConfig()
	if err == nil { // file exists
		// writes in new defaults, but keeps from overwriting user set configs
		// since config keys might be added often in early stages of this project
		if err := saveMissingDefaults(); err != nil {
			return fmt.Errorf("migrating missing config keys | %w", err)
		}
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
	V.SetDefault(KeyPathPluginsDir, filepath.Join(defaultConfigDir, "plugins"))
	V.SetDefault(KeyPathLogsDir, filepath.Join(home, ".local", "state", constant.APP_NAME, "plugins.log"))

	V.SetDefault(KeyDefaultFilterGroup, "")
	V.SetDefault(KeyDefaultFilterTag, "")
	V.SetDefault(KeyDefaultReviewMode, "")
	V.SetDefault(KeyDefaultReviewRenderer, "")
	V.SetDefault(KeyDefaultReviewLimit, -1)
	return nil
}

func SetUserDefault(key string, value interface{}) error {
	V.Set(key, value)
	if err := V.WriteConfig(); err != nil {
		return fmt.Errorf("persisting new config value | %w", err)
	}

	return nil
}

// Resolve evaluates a hierarchical chain of values to find the active configuration setting.
// It prioritizes an explicit runtime override value (e.g., from a user's CLI flag). If that override
// is empty or set to "default", it pulls the value matching the specified key from the configuration file.
// If the file configuration is also missing, empty, or set to default, it returns the hardcoded system fallback string.
func Resolve(key string, override string, systemFallback string) string {
	overrideTrimmed := strings.ToLower(strings.TrimSpace(override))

	if overrideTrimmed != "" && overrideTrimmed != "default" {
		return override
	}

	cfgVal := V.GetString(key)
	cfgValTrimmed := strings.ToLower(strings.TrimSpace(cfgVal))

	if cfgValTrimmed != "" && cfgValTrimmed != "default" {
		return cfgVal
	}

	return systemFallback
}

func saveMissingDefaults() error {
	hasChanges := false

	for key, value := range V.AllSettings() {
		if !V.InConfig(key) {
			V.Set(key, value)
			hasChanges = true
		}
	}

	if hasChanges {
		if err := V.WriteConfig(); err != nil {
			return fmt.Errorf("writing updated config with defaults | %w", err)
		}
	}
	return nil
}
