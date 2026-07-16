package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

type OSFlag uint8

const (
	LINUX OSFlag = 1 << iota
	WINDOWS
	MACOS
	UNKNOWN
)

var platform OSFlag

func init() {
	setOS()
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

func DataDirectory() (string, error) {
	var result string
	switch {
	case IsLinux():
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			result = filepath.Join(home, ".local", "share")
		} else {
			result = dataDir
		}
	default:
		return os.UserConfigDir()
	}

	return result, nil
}
