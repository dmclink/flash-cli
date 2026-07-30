package utils

import (
	"os"
	"os/exec"
	"strings"

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

func SplitFieldsAndCommas(s string) []string {
	intermediate := strings.Fields(s)
	result := []string{}
	for _, i := range intermediate {
		result = append(result, SplitAtCommas(i)...)
	}
	return result
}

// splitAtCommas delimits a string by commas and removes duplicate values
//
// NOTE: order is not maintained
func SplitAtCommas(s string) []string {
	sp := strings.Split(s, ",")

	m := make(map[string]bool, len(sp))
	for _, ss := range sp {
		m[ss] = true
	}

	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	return result
}
