package cmd

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/edit"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/spf13/cobra"
)

func NewEditCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "edit",
		Short:              "Opens up default editor to edit cards selected by filters",
		Long:               editLongString,
		DisableFlagParsing: true,
		Annotations: map[string]string{
			"filter": "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(a.Args.Filters) == 0 {
				return fmt.Errorf("Requires at least one filter.")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO:
			ctx := cmd.Context()
			filters := parser.ParseSearchFilters(a.Config.V, a.Args, false)
			cards, err := database.GetFlashcards(a.DB, filters)
			if err != nil {
				return fmt.Errorf("getting flashcards from db | %w", err)
			}

			if len(cards) == 0 {
				fmt.Println("Nothing in this query to edit. Check for typos or change filters")
				return nil
			}

			edit.Edit(ctx, a, cards)
			return nil
		},
	}

	cmd.SetHelpTemplate(editHelpTemplate)
	cmd.SetUsageTemplate(universalUsageTemplate)
	return cmd
}

var editLongString = `If more than one card is selected, opens them up one by one. If the editor is closed 
without saving or no changes made, skips editing the card. If some syntax error or 
invalid data is entered in your edits, reopens the editor in its original state.

Fields marked with # comments are not allowed to be changed.

Once changes are made, saved to buffer, and the editor is closed, the new 
edits are persisted to the database`

var editHelpTemplate = `NAME
  flash-cli {{.Name}} - {{.Short}}

USAGE
  flash-cli {{if .Annotations.filter}}<filter> {{else}}         {{end -}} {{.Name}} {{if .Annotations.modsyntax}}{{.Annotations.modsyntax}}{{end}}

DESCRIPTION
{{.Long}}

CONFIGURATION
  default.editor
    The name or path of the text editor executable (e.g., "nvim", "nano"). 
    If unset, falls back to the system's $VISUAL or $EDITOR environment 
    variables, or defaults to an available system editor.

FILTERS
  Requires at least one filter to select a card. Doesn't allow default filter configurations to apply.

MODS
  Mods are silently ignored.

EXAMPLES
  flash-cli 1           edit      Opens an editor for card with id=1
  flash-cli group:foo   edit      Opens an editor for each card in group 'foo'
  flash-cli 4,5,10      edit      Opens an editor for cards with ids 4, 5, and 10
`
