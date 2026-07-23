package cmd

import (
	"fmt"

	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/cobra"
)

func NewVersionCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the flash-cli version",
		Long:  "Prints the flash-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(constant.VERSION)
		},
	}
}
