package logger

import (
	"fmt"
	"os"

	"github.com/dmclink/flash-cli/internal/config"
	"github.com/hashicorp/go-hclog"
)

func InitPluginLogger(cfg *config.Config) (hclog.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil cfg passed to plugin logger")
	}
	if cfg.V == nil {
		return nil, fmt.Errorf("nil viper passed to config")
	}
	logDir := cfg.V.GetString(config.KeyPathLogsDir)
	logFile, err := os.OpenFile(logDir, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:   "plugin-logger",
		Output: logFile,
		Level:  hclog.Warn,
	}), nil
}
