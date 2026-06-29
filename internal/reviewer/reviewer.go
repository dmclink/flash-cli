package reviewer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/utils"
)

func Review(cards []database.Flashcard) error {
	if len(cards) == 0 {
		fmt.Println("Nothing in this query to review. Try changing review settings or filters")
		return nil
	}

	// alternate terminal buffer since we use clears below, might not be supported
	// in early win10 and earlier but that's their problem
	fmt.Print("\033[?1049h")
	os.Stdout.Sync()
	defer func() {
		fmt.Print("\033[?1049l")
		os.Stdout.Sync()
	}()

	err := utils.ClearScreen()
	if err != nil {
		return err
	}

	for _, card := range cards {
		fmt.Println("Front:")
		fmt.Println(card.Front)
		fmt.Println()
		fmt.Println("Press enter to flip card...")

		_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("failed capturing keystore for flipping card | %w", err)
		}

		err = utils.ClearScreen()
		if err != nil {
			return err
		}

		fmt.Println("Front:")
		fmt.Println(card.Front)
		fmt.Println()

		fmt.Println("Back:")
		fmt.Println(card.Back)
		fmt.Println()
		fmt.Println("Press enter for next card...")

		_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("failed capturing keystore for flipping card | %w", err)
		}

		err = utils.ClearScreen()
		if err != nil {
			return err
		}
	}

	return nil
}
