package cmd

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/cobra"
)

func NewVersionCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the flash-cli version",
		Long:  "Prints the flash-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(constant.VERSION)
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Short)
		fmt.Println()
		fmt.Println(versionHelpStr)
	})

	return cmd
}

var versionHelpStr = `USAGE
  flash-cli version

FILTERS
  Silently ignores all filters

MODS
  Silently ignores all mods`
