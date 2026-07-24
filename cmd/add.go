package cmd

import (
	"fmt"
	"strings"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/config"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
)

func NewAddCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new flashcard",
		Annotations: map[string]string{
			"modsyntax": "<mods>",
			"filter":    "true",
		},
		DisableFlagParsing: true,
		// TODO: do i put the usage here? explain which filters work ie. IDs and UUIDS are not available, -tags are not available
		Long: "Adds new flashcard. The front and back of the flashcard is input to <mods> and can be either space separated values or a double quoted string. <mods> must include delimiter to distinguish between front and back or throws error.\nOnly group type <filters> are allowed to designate which groups the flashcard belongs to.\nNew flashcards have a default last_reviewed set to the time of creation.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			mods := strings.Join(a.Args.Mods, " ")
			delim := a.Config.V.GetString(config.KeyAddSeparator)
			if !strings.Contains(mods, delim) {
				return fmt.Errorf("<mods> for `add` command must contain delimiter '%s'", delim)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mods := strings.Join(a.Args.Mods, " ")
			delim := a.Config.V.GetString(config.KeyAddSeparator)
			splitMods := strings.SplitN(mods, delim, 2)
			if len(splitMods) < 2 {
				return fmt.Errorf("something went wrong, delimiter doesn't exist but should have been verified in PreRun")
			}

			front := splitMods[0]
			back := splitMods[1]

			filters := parser.ParseSearchFilters(a.Args)

			// TODO: extract plugin data from filters and pass them to AddFlashcard

			err := database.AddFlashcard(a.DB, front, back, filters.Groups, filters.Tags)
			if err != nil {
				return err
			}

			var output strings.Builder
			output.WriteString("Added 1 new flashcard")
			if len(filters.Groups) > 0 {
				groups := make([]string, 0, len(filters.Groups))
				for _, g := range filters.Groups {
					groups = append(groups, g.Value)
				}
				output.WriteString(" to group(s): ")
				output.WriteString(strings.Join(groups, ", "))
			}
			fmt.Println(output.String())

			return nil
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Short)
		fmt.Println()
		fmt.Println(addHelpStr)
	})

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		// cobra automatically prints the error string here
		fmt.Println()
		fmt.Println(addUsageStr)
		return nil
	})
	return cmd
}

var addHelpStr = `USAGE
  flash-cli <filters> add <mods>    Adds a new card with <filters> groups and tags and <mods> data

FILTERS
  Group and +tag filters will be applied to the new card. IDs, UUIDs, and -tag filters will be ignored.

MODS
  At least one mod is required and at least one must contain the card separator symbol.
  The card separator is set to '::' by default, this can be changed with the config command.
  All mods following the add command until the first encountered card separator symbol will be the new card's front.
  All remaining mods after the separator will be the card's back. It is highly recommended to wrap mods in 
  single quotes to avoid bash errors when using symbols or entering multi line cards.

EXAMPLES
Creates a new card in group 'foo'. Where Front='this is a new card' and Back='and this is its back'
  flash-cli group:foo add this is a new card::and this is its back

Creates a new card with multiline Front text and attaches tag 'mc':
  flash-cli +mc add 'Pick your favorite var:
a. Foo
b. Bar
c. Baz::The answer is c'`

var addUsageStr = `SYNTAX ERROR
  Missing mods or card separator.

USAGE
  flash-cli <filters> add <mods>
			
Run 'flash-cli help add' for a detailed list of available options`
