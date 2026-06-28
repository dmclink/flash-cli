package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dmclink/flash-cli/internal/constant"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewAddCmd(db *sql.DB, v *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "add [mods]",
		Short: "Add new flashcard",
		Long:  "Adds new flashcard. The front and back of the flashcard is input to <mods> and can be either space separated values or a double quoted string. <mods> must include delimiter to distinguish between front and back or throws error.\nOnly group type <filters> are allowed to designate which groups the flashcard belongs to.\nNew flashcards have a default last_reviewed set to the time of creation.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			parsedArgs, err := utils.GetParsedArgs(cmd)
			if err != nil {
				return err
			}

			mods := strings.Join(parsedArgs.Mods, " ")
			delim := v.GetString(constant.VIPER_KEY_DELIMITER)
			if !strings.Contains(mods, delim) {
				return fmt.Errorf("<mods> for `add` command must contain delimiter '%s'", delim)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedArgs, err := utils.GetParsedArgs(cmd)
			if err != nil {
				return err
			}

			mods := strings.Join(parsedArgs.Mods, " ")
			delim := v.GetString(constant.VIPER_KEY_DELIMITER)
			splitMods := strings.SplitN(mods, delim, 2)
			if len(splitMods) < 2 {
				return fmt.Errorf("something went wrong, delimiter doesn't exist but should have been verified in PreRun")
			}

			front := splitMods[0]
			back := splitMods[1]

			err = database.AddFlashcard(db, front, back)
			if err != nil {
				return err
			}

			// TODO: add some statement here to say what groups it was added after groups implemented
			fmt.Println("Added 1 new flashcard")

			return nil
		},
	}
}
