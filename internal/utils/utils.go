package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
)

func GetParsedArgs(cmd *cobra.Command) (parser.ParsedArgs, error) {
	parsedArgs, ok := cmd.Context().Value(constant.PARSED_ARGS_KEY).(parser.ParsedArgs)
	if !ok {
		return parser.ParsedArgs{}, fmt.Errorf("failed to cast ParsedArgs")
	}

	return parsedArgs, nil
}

// TODO: only works for linux terminals, future support for other OS by saving their clear func in a map and lookup os
func ClearScreen() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
