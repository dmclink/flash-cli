package cmd

import (
	"fmt"
	"strings"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/spf13/cobra"
)

var helpStr = `USAGE
  flash-cli config <key> <value>  Set or change a configuration option in config file
  flash-cli config <key>          Remove a custom entry from config file (falls back to defaults)
FILTERS
  Filters for this command are silently ignored.
MODS
  The first mod parsed is the config key that will be added/modified in the file.
  All remaining mods will be treated as a single string white whitespace collapsed but preserved.
  ie. 'flash-cli config foo bar    baz' will set the 'foo' key to 'bar baz'`

var usageStr = `SYNTAX ERROR
  Invalid configuration syntax provided.
USAGE
  flash-cli config <key> <value>
  flash-cli config <key>
			
Run 'flash-cli help config for a detailed list of available options'`

func NewConfigCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "config",
		Short:              "Alters the values in the config file",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(a.Args.Mods) == 0 {
				return fmt.Errorf("Specify the name of the config variable to modify.")
			}

			cfgKey := a.Args.Mods[0]

			ctx := cmd.Context()
			if len(a.Args.Mods) == 1 {
				// TODO: call config.Remove here after it is implemented
				return a.Config.Remove(ctx, cfgKey)
			}

			return a.Config.Set(ctx, cfgKey, strings.Join(a.Args.Mods[1:], " "))
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(helpStr)
	})

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Println(usageStr)
		return nil
	})

	return cmd
}
