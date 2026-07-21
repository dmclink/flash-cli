package logger

import (
	"io"
	"os"
	"path/filepath"

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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, ".local", "state", "flash-cli")
	_ = os.MkdirAll(logDir, 0o755)

	logPath := filepath.Join(logDir, "plugins.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
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
