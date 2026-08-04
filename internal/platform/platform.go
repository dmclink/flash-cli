package platform

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/dmclink/flash-cli/internal/constant"
)

type OSFlag uint8

const (
	LINUX OSFlag = 1 << iota
	WINDOWS
	MACOS
	UNKNOWN
)

var (
	platform OSFlag

	dataDir    string
	pluginsDir string
	logsDir    string
	configsDir string
)

func Init() error {
	setOS()

	// order matters for these
	err := setConfigsDir()
	if err != nil {
		return err
	}
	err = setDataDir()
	if err != nil {
		return err
	}
	err = setPluginsDir()
	if err != nil {
		return err
	}
	err = setLogsDir()
	if err != nil {
		return err
	}
	return nil
}

func IsLinux() bool {
	return platform&LINUX != 0
}

func IsWindows() bool {
	return platform&WINDOWS != 0
}

func IsMacOS() bool {
	return platform&MACOS != 0
}

func ConfigsDir() string {
	return configsDir
}

func DataDir() string {
	return dataDir
}

func LogsDir() string {
	return logsDir
}

func PluginsDir() string {
	return pluginsDir
}

func EditorFallback() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	switch {
	case IsLinux(), IsMacOS():
		return "vi"
	case IsWindows():
		return "notepad"
	default:
		return ""
	}
}

func setOS() {
	switch runtime.GOOS {
	case "linux":
		platform |= LINUX
	case "windows":
		platform |= WINDOWS
	case "darwin":
		platform |= MACOS
	default:
		platform |= UNKNOWN
	}
}

func setPluginsDir() error {
	dataDir := DataDir()
	result := filepath.Join(dataDir, "plugins")

	err := os.MkdirAll(result, 0o755)
	if err != nil {
		return err
	}

	pluginsDir = result
	return nil
}

func setLogsDir() error {
	var result string
	switch {
	case IsLinux():
		logDir := os.Getenv("XDG_CACHE_HOME")
		if logDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			result = filepath.Join(home, ".cache")
		} else {
			result = logDir
		}
	case IsMacOS():
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		result = filepath.Join(home, "Library", "Logs")
	default:
		result = ConfigsDir()
	}

	result = filepath.Join(result, constant.APP_NAME)

	err := os.MkdirAll(result, 0o755)
	if err != nil {
		return err
	}

	logsDir = result
	return nil
}

func setConfigsDir() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	result := filepath.Join(dir, constant.APP_NAME)

	err = os.MkdirAll(result, 0o755)
	if err != nil {
		return err
	}
	configsDir = result
	return nil
}

func setDataDir() error {
	var result string
	switch {
	case IsLinux():
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			result = filepath.Join(home, ".local", "share")
		} else {
			result = dataDir
		}
	default:
		result = ConfigsDir()
	}

	result = filepath.Join(result, constant.APP_NAME)

	err := os.MkdirAll(result, 0o755)
	if err != nil {
		return err
	}

	dataDir = result
	return nil
}
