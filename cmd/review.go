package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/ext"
	"github.com/dmclink/flash-cli/internal/parser"
	"github.com/dmclink/flash-cli/internal/utils"
	"github.com/spf13/cobra"
)

func NewReviewCmd(db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:                "review",
		Short:              "Review flashcards",
		DisableFlagParsing: true,
		// TODO: change these comments about mods, filters, and config after those are implemented
		Long: "Review flashcards in order by set by mods or defaults ordered by last reviewed, oldest first. Shows one flashcard at a time. Can be filtered by groups or ID ranges. Settings can be changed with config",
		// TODO: parse filters and mods here
		// PreRunE: func(cmd *cobra.Command, args []string) error {
		// 	return nil
		// },
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			parsedArgs, err := parser.ExtractParsedArgs(cmd)
			if err != nil {
				return fmt.Errorf("extracting parsed args | %w", err)
			}

			filters := parser.ParseSearchFilters(parsedArgs)
			cards, err := database.GetFlashcards(db, filters)
			if err != nil {
				return fmt.Errorf("getting flashcards from db | %w", err)
			}

			// TODO:
			// limit the card based on limit filter

			if len(cards) == 0 {
				fmt.Println("Nothing in this query to review. Try changing review settings or filters")
				return nil
			}

			reviewMode, unparsedMods := removeMode(parsedArgs.Mods)

			processor, procCleanup, err := ext.DispenseReviewProcessor(reviewMode)
			if err != nil {
				return fmt.Errorf("dispensing review processor | %w", err)
			}
			defer procCleanup()

			cards, err = processor.Process(ctx, cards, unparsedMods)
			if err != nil {
				return fmt.Errorf("processing cards | %w", err)
			}

			rendererString, unparsedMods := removeRenderer(unparsedMods)

			renderer, renderCleanup, err := ext.DispenseRenderer(rendererString)
			if err != nil {
				return fmt.Errorf("dispensing renderer | %w", err)
			}
			defer renderCleanup()

			// alternate terminal buffer since we use clears below, might not be supported
			// in early win10 and earlier but that's their problem
			fmt.Print("\033[?1049h")
			os.Stdout.Sync()
			defer func() {
				fmt.Print("\033[?1049l")
				os.Stdout.Sync()
			}()

			err = utils.ClearScreen()
			if err != nil {
				return err
			}

			inputChan := make(chan string)
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					inputChan <- scanner.Text()
				}
			}()

			startupBanner, frontInstruction, backInstruction, err := renderer.Init(ctx)
			if err != nil {
				return err
			}
			frontInstruction = frontInstructionFallback(frontInstruction)
			backInstruction = backInstructionFallback(backInstruction)

			if startupBanner != "" {
				utils.ClearScreen()
				fmt.Println(startupBanner)
				waitForInput(ctx, inputChan)
			}

			cardCount := len(cards)
			for i, card := range cards {
				cardNum := i + 1
				front, back, progress, err := renderer.Render(ctx, card, cardNum, cardCount, unparsedMods)
				if err != nil {
					if ctx.Err() != nil {
						break
					}
					return fmt.Errorf("running renderer | %w", err)
				}

				progress = progressFallback(progress, cardNum, cardCount)

				utils.ClearScreen()
				fmt.Println(front)
				printLockedFooter(progress, frontInstruction)

				if !waitForInput(ctx, inputChan) {
					break
				}

				utils.ClearScreen()
				fmt.Println(back)

				printLockedFooter(progress, backInstruction)

				if !waitForInput(ctx, inputChan) {
					break
				}
			}
			if ctx.Err() != nil {
				fmt.Println("\nReview session aborted by user.")
			} else {
				fmt.Println("All cards complete")
			}
			return nil
		},
	}
}

func fallbackStr(str string, fallback string) string {
	if str == "" {
		return fallback
	}
	return str
}

func frontInstructionFallback(instruction string) string {
	return fallbackStr(instruction, fmt.Sprintf("%sPress [ENTER] to reveal the answer... ('q' to quit): ", Yellow))
}

func backInstructionFallback(instruction string) string {
	return fallbackStr(instruction, fmt.Sprintf("%sPress [ENTER] for the next card... ('q' to quit): ", Yellow))
}

func progressFallback(progress string, cardNum int, cardCount int) string {
	return fallbackStr(progress, fmt.Sprintf("%s[%d/%d]", Yellow, cardNum, cardCount))
}

const (
	Reset     = "\033[0m"
	Yellow    = "\033[33m"
	ClearLine = "\033[K" // Wipes out old text to prevent text ghosting
)

// SetTerminalMargins locks the bottom row of the screen.
// It tells the terminal: "Let lines 1 through (Bottom-1) scroll, but lock the last line."
func setTerminalMargins() {
	fmt.Print("\033[1;997r")
}

// ResetTerminalMargins restores standard fullscreen scrolling behavior.
func resetTerminalMargins() {
	fmt.Print("\033[r")
}

// printLockedFooter locks the progress bar to row 998 and the prompt to row 999
func printLockedFooter(progress, instruction string) {
	fmt.Printf("\033[999;1H\033[1A%s%s%s", ClearLine, progress, Reset)
	fmt.Printf("\033[999;1H%s%s%s", ClearLine, instruction, Reset)
}

func waitForInput(ctx context.Context, inputChan <-chan string) bool {
	select {
	case <-ctx.Done():
		return false
	case input, ok := <-inputChan:
		if !ok {
			return false // Channel closed, stop
		}

		exitSignals := []string{"q", "quit", "exit"}
		if slices.Contains(exitSignals, input) {
			return false // User requested exit, stop
		}

		return true
	}
}

func hasModPrefix(s, name string) bool {
	nameWithColon := name + ":"
	nameWithEqual := name + "="
	return strings.HasPrefix(s, nameWithColon) || strings.HasPrefix(s, nameWithEqual)
}

func findPrefixIdx(mods []string, prefix string) int {
	for i, mod := range mods {
		if hasModPrefix(mod, prefix) {
			return i
		}
	}
	return -1
}

func stripModPrefix(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' || s[i] == ':' {
			return s[i+1:]
		}
	}
	return s
}

func removeFromMods(mods []string, modName string) (string, []string) {
	idx := findPrefixIdx(mods, modName)
	result := ""
	for idx != -1 {
		result = stripModPrefix(mods[idx])
		mods = append(mods[:idx], mods[idx+1:]...)
		idx = findPrefixIdx(mods, modName)
	}

	return result, mods
}

func removeRenderer(mods []string) (string, []string) {
	return removeFromMods(mods, "renderer")
}

func removeMode(mods []string) (string, []string) {
	return removeFromMods(mods, "mode")
}
