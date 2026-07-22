package logger

import (
	"io"
	"os"

	"github.com/dmclink/flash-cli/internal/config"
	"github.com/hashicorp/go-hclog"
)

// L is the global Logger instance available to entire host codebase.
// Initialized to discard for unit testing
var L hclog.Logger = hclog.New(&hclog.LoggerOptions{
	Name:   "discard",
	Output: io.Discard,
	Level:  hclog.Off,
})

func InitPluginLogger() error {
	logDir := config.V.GetString(config.KeyPathLogsDir)
	logFile, err := os.OpenFile(logDir, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	L = hclog.New(&hclog.LoggerOptions{
		Name:   "plugin-logger",
		Output: logFile,
		Level:  hclog.Warn,
	})
	return nil
}
