package cmd

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
)

var addCommand = &cobra.Command{
	Use:   "add [mods]",
	Short: "Add new flashcard",
	Long:  "Adds new flashcard. The front and back of the flashcard is input to <mods> and can be either space separated values or a double quoted string. <mods> must include delimiter to distinguish between front and back or throws error.\nOnly group type <filters> are allowed to designate which groups the flashcard belongs to.\nNew flashcards have a default last_reviewed set to the time of creation.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		parsedArgs, ok := cmd.Context().Value(constant.PARSED_ARGS_KEY).(parser.ParsedArgs)
		if !ok {
			return fmt.Errorf("failed to cast ParsedArgs")
		}
		// TODO: check for card delimiter in mods
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedArgs, ok := cmd.Context().Value(constant.PARSED_ARGS_KEY).(parser.ParsedArgs)
		if !ok {
			return fmt.Errorf("failed to cast ParsedArgs")
		}

		return nil
	},
}
