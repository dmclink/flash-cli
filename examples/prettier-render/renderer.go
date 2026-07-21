package main

import (
	"fmt"
	"os"
	"strings"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	"golang.org/x/term"
)

// UI Palette Color Codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Dim    = "\033[2m"
	BgLine = "\033[48;5;235m"
)

// some static metadata variables that are sent on plugin connect
var startupBanner = `
   ___          _   _   _            __                _           
  / _ \_ __ ___| |_| |_(_) ___ _ __ /__\ ___ _ __   __| | ___ _ __ 
 / /_)/ '__/ _ \ __| __| |/ _ \ '__/ \/// _ \ '_ \ / _| |/ _ \ '__|
/ ___/| | |  __/ |_| |_| |  __/ | / _  \  __/ | | | (_| |  __/ |   
\/    |_|  \___|\__|\__|_|\___|_| \/ \_/\___|_| |_|\__,_|\___|_|   
	`

var (
	instructionFront = "Enter to flip, q to quit!"
	instructionBack  = "Enter to next, q to quit!"
)

// DESIGN CONTROLLER:
// Edit this struct and its 'RenderCard' method to customize your layouts.
// This method is called sequentially by main.go every time a card is drawn.
type PrettierRenderer struct{}

// RenderCard processes a card and outputs your terminal styling format designs.
// Returns exactly three strings: (front_view, back_view, progress_bar)
func (r *PrettierRenderer) RenderCard(card *common.Flashcard, cardNum, totalCards int32, modifiers []string) (string, string, string) {
	// Query user's current terminal width dynamically to keep split grids responsive
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth < 40 {
		termWidth = 80 // Safe fallback scale boundary width
	}

	targetGridWidth := (termWidth * 3) / 4
	if targetGridWidth < 60 {
		targetGridWidth = 60 // Enforce a sane minimum horizontal scale
	}

	// 1. Generate full layout structures for both presentation states
	frontView := generateSplitView(card, false, targetGridWidth)
	// NOTE: backView likely also contains the full layout of frontView
	// unless you don't want the question portion of the card visible after user "flips"
	backView := generateSplitView(card, true, targetGridWidth)

	// 2. Format a matching text-based timeline progress block
	// feel free to skip this or return empty "" to fall back to a
	// default progress indicator like [1/5] or pass a non empty
	// string like " " to remove progress altogether
	progressBar := fmt.Sprintf("%s[Progress Metrics: %d/%d]%s", Dim, cardNum, totalCards, Reset)

	return frontView, backView, progressBar
}

// wrapText breaks a string down into lines that perfectly fit a specific width without breaking words
func wrapText(text string, maxWidth int) []string {
	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		words := strings.Fields(strings.TrimSpace(para))
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= maxWidth {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}
	return lines
}

func generateSplitView(card *common.Flashcard, showBack bool, gridWidth int) string {
	var sb strings.Builder

	colWidth := (gridWidth / 2) - 3

	frontHeader := "FRONT (QUESTION)"
	backHeader := "BACK (ANSWER)"
	sb.WriteString(fmt.Sprintf("%s%s %-*s │ %-*s %s\n", BgLine, Bold, colWidth+1, frontHeader, colWidth+1, backHeader, Reset))

	dividerLine := strings.Repeat("─", colWidth+2)
	sb.WriteString(fmt.Sprintf("%s%s%s─┼%s%s\n", Dim, Cyan, dividerLine, dividerLine, Reset))

	frontLines := wrapText(card.Front, colWidth)
	backLines := []string{}
	if showBack {
		backLines = wrapText(card.Back, colWidth)
	}

	maxLines := len(frontLines)
	if len(backLines) > maxLines {
		maxLines = len(backLines)
	}

	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(frontLines) {
			left = frontLines[i]
		}
		if showBack && i < len(backLines) {
			right = backLines[i]
		}

		sb.WriteString(fmt.Sprintf(" %-*s %s│%s %s%-*s%s\n",
			colWidth+1, left,
			Dim+Cyan, Reset,
			Green, colWidth+1, right, Reset,
		))
	}

	bottomLine := strings.Repeat("─", colWidth+2)
	sb.WriteString(fmt.Sprintf("%s%s%s─┴%s%s\n", Dim, Cyan, bottomLine, bottomLine, Reset))

	return sb.String()
}
