package cmd

import (
	"database/sql"
	"fmt"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewVersionCmd(db *sql.DB, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the flash-cli version",
		Long:  "Prints the flash-cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(constant.VERSION)
		},
	}
}
