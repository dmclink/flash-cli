package cmd

import (
	"fmt"
	"strings"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/spf13/cobra"
)

func NewConfigCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Alters set values in the config file",
		Annotations: map[string]string{
			"modsyntax": "[name [value | '']]",
		},
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(a.Args.Mods) == 0 {
				return fmt.Errorf("Specify the name of the config variable to modify.")
			}

			cfgKey := a.Args.Mods[0]

			ctx := cmd.Context()
			if len(a.Args.Mods) == 1 {
				return a.Config.Remove(ctx, cfgKey)
			}

			return a.Config.Set(ctx, cfgKey, strings.Join(a.Args.Mods[1:], " "))
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Short)
		fmt.Println()
		fmt.Println(cfgHelpStr)
	})

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		// cobra automatically prints the error string here
		fmt.Println()
		fmt.Println(cfgUsageStr)
		return nil
	})

	return cmd
}

var cfgHelpStr = `USAGE
  flash-cli config <key> <value>  Set or change a configuration option in config file
  flash-cli config <key>          Remove a custom entry from config file (falls back to defaults)

FILTERS
  Filters for this command are silently ignored.

MODS
  At least one mod is required.
  The first mod parsed is the config key that will be added/modified in the file.
  All remaining mods will be treated as a single string white whitespace collapsed but preserved.

EXAMPLES
  flash-cli config foo bar           Sets the 'foo' key to 'bar'
  flash-cli config foo -1            Sets the 'foo' key to -1
  flash-cli config foo true          Sets the 'foo' key to true
  flash-cli config foo bar    baz    Sets the 'foo' key to 'bar baz'
  flash-cli config foo               Deletes the 'foo' key from the config file`

var cfgUsageStr = `SYNTAX ERROR
  Invalid configuration syntax provided.

USAGE
  flash-cli config <key> <value>
  flash-cli config <key>
			
Run 'flash-cli help config for a detailed list of available options`
