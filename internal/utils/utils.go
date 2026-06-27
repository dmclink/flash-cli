package utils

import (
	"fmt"

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
