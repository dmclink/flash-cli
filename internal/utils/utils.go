package utils

import (
	"os"
	"os/exec"

	"github.com/google/uuid"
)

// TODO: only works for linux terminals, future support for other OS by saving their clear func in a map and lookup os
func ClearScreen() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// IsValidUUID returns true if s is a valid UUID otherwise returns false
func IsValidUUID(s string) bool {
	if err := uuid.Validate(s); err == nil {
		return true
	}

	return false
}
