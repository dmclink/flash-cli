package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/viper"
)

type Config struct {
	V *viper.Viper
}

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

func InitConfig() (*Config, error) {
	v := viper.New()
	err := setInitDefaults(v)
	if err != nil {
		return nil, fmt.Errorf("setting config defaults | %w", err)
	}

	var configPath string

	if envPath := os.Getenv("FLASH_CLI_CONFIG_DIR"); envPath != "" {
		configPath = envPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to find home directory for os | %w", err)
		}
		configPath = filepath.Join(home, ".config", constant.APP_NAME)
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.SetEnvPrefix("FLASH_CLI")
	v.AutomaticEnv()

	err = v.ReadInConfig()
	result := &Config{v}
	if err == nil { // file exists
		return result, nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		if err := os.MkdirAll(configPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating config directory | %w", err)
		}
		if err := v.SafeWriteConfig(); err != nil {
			return nil, fmt.Errorf("creating default config | %w", err)
		}
		return result, nil
	}

	return result, fmt.Errorf("failed reading config file | %w", err)
}

func setInitDefaults(v *viper.Viper) error {
	v.SetDefault(KeyAddSeparator, "::")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to find home directory for os | %w", err)
	}
	defaultConfigDir := filepath.Join(home, ".config", constant.APP_NAME)
	v.SetDefault(KeyPathPluginsDir, filepath.Join(defaultConfigDir, "plugins"))
	v.SetDefault(KeyPathLogsDir, filepath.Join(home, ".local", "state", constant.APP_NAME, "plugins.log"))

	v.SetDefault(KeyDefaultFilterGroup, "")
	v.SetDefault(KeyDefaultFilterTag, "")
	v.SetDefault(KeyDefaultReviewMode, "")
	v.SetDefault(KeyDefaultReviewRenderer, "")
	v.SetDefault(KeyDefaultReviewLimit, -1)
	return nil
}

func (c *Config) Set(ctx context.Context, key string, value string) error {
	confirmationMessage := fmt.Sprintf("Are you sure you want to add '%s' with a value of '%v'? (yes/no) ", key, value)
	resultMessage := fmt.Sprintf("Config file %s modified.", c.V.ConfigFileUsed())
	rejectMessage := "No changes made."
	exists := c.V.InConfig(key)
	if exists {
		prev := c.V.Get(key)
		confirmationMessage = fmt.Sprintf("Are you sure you want to change the value of '%s' from '%v' to '%v'? (yes/no) ", key, prev, value)
	}

	inputChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
	}()

	positive := []string{"y", "ye", "yes"}
	negative := []string{"n", "no"}
loop:
	for {
		fmt.Print(confirmationMessage)
		select {
		case <-ctx.Done():
			break loop
		case rawInput := <-inputChan:
			input := strings.ToLower(rawInput)
			if slices.Contains(negative, input) {
				fmt.Println(rejectMessage)
				return nil
			}
			if slices.Contains(positive, input) {
				break loop
			}
			continue
		}
	}

	c.setTypedConfigValue(key, value)
	if err := c.V.WriteConfig(); err != nil {
		return fmt.Errorf("persisting new config value | %w", err)
	}

	fmt.Println(resultMessage)
	return nil
}

func getConfigFileSettings(c *Config) (map[string]interface{}, error) {
	fileOnlyViper := viper.New()

	configFile := c.V.ConfigFileUsed()
	if configFile == "" {
		return nil, fmt.Errorf("no config file is currently loaded")
	}
	fileOnlyViper.SetConfigFile(configFile)

	if err := fileOnlyViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file isolation copy | %w", err)
	}

	return fileOnlyViper.AllSettings(), nil
}

func (c *Config) Remove(ctx context.Context, key string) error {
	cfgMap, err := getConfigFileSettings(c)
	if err != nil {
		return err
	}

	exists := c.V.InConfig(key)
	if !exists {
		fmt.Printf("No entry named '%s' found.\n", key)
		return nil
	}

	confirmationMessage := fmt.Sprintf("Are you sure you want to remove '%s'? (yes/no) ", key)
	resultMessage := fmt.Sprintf("Config file %s modified.", c.V.ConfigFileUsed())
	rejectMessage := "No changes made."

	inputChan := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
	}()

	positive := []string{"y", "ye", "yes"}
	negative := []string{"n", "no"}
loop:
	for {
		fmt.Print(confirmationMessage)
		select {
		case <-ctx.Done():
			break loop
		case rawInput := <-inputChan:
			input := strings.ToLower(rawInput)
			if slices.Contains(negative, input) {
				fmt.Println(rejectMessage)
				return nil
			}
			if slices.Contains(positive, input) {
				break loop
			}
			continue
		}
	}
	keys := strings.Split(key, ".")
	m := cfgMap
	for i := 0; i < len(keys)-1; i++ {
		m = m[keys[i]].(map[string]any)
	}
	delete(m, keys[len(keys)-1])

	cfgFile := c.V.ConfigFileUsed()
	V := viper.New()
	V.SetConfigFile(cfgFile)

	for k, v := range cfgMap {
		V.Set(k, v)
	}

	if err := V.WriteConfig(); err != nil {
		return fmt.Errorf("persisting new config value | %w", err)
	}

	fmt.Println(resultMessage)
	return nil
}

// Resolve evaluates a hierarchical chain of values to find the active configuration setting.
// It prioritizes an explicit runtime override value (e.g., from a user's CLI flag). If that override
// is empty or set to "default", it pulls the value matching the specified key from the configuration file.
// If the file configuration is also missing, empty, or set to default, it returns the hardcoded system fallback string.
func (c *Config) Resolve(key string, override string, systemFallback string) string {
	overrideTrimmed := strings.ToLower(strings.TrimSpace(override))

	if overrideTrimmed != "" && overrideTrimmed != "default" {
		return override
	}

	cfgVal := c.V.GetString(key)
	cfgValTrimmed := strings.ToLower(strings.TrimSpace(cfgVal))

	if cfgValTrimmed != "" && cfgValTrimmed != "default" {
		return cfgVal
	}

	return systemFallback
}

// SetTypedConfigValue parses a raw string CLI argument into its most appropriate
// native Go type (int, bool, or string) and assigns it to the specified Viper key.
// It modifies the active in-memory Viper registry but does not commit changes to
// the physical disk.
func (c *Config) setTypedConfigValue(key string, rawValue string) {
	var typedValue interface{} = rawValue

	if intVal, err := strconv.Atoi(rawValue); err == nil {
		typedValue = intVal
	} else if boolVal, err := strconv.ParseBool(rawValue); err == nil {
		typedValue = boolVal
	}

	c.V.Set(key, typedValue)
}
