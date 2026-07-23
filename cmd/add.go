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
		Use:                "add",
		Short:              "Add new flashcard",
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

	return cmd
}
