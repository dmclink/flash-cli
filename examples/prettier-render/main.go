package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	"github.com/dmclink/flash-cli/shared"
	"github.com/hashicorp/go-plugin"
	"golang.org/x/term"
)

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Dim    = "\033[2m"
	BgLine = "\033[48;5;235m"
)

func main() {
	renderImpl := &RenderHandler{}

	pluginMap := map[string]plugin.Plugin{
		shared.CAPABILITY_RENDER: &shared.GenericGRPCPlugin[
			shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse],
			shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]]{
			Impl:               renderImpl,
			RegisterServerFunc: shared.PluginMap[shared.CAPABILITY_RENDER].(*shared.GenericGRPCPlugin[shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse], shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]]).RegisterServerFunc,
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

func ClearScreen() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type RenderHandler struct{}

func (h *RenderHandler) Process(ctx context.Context, req *render.ProcessRequest) (*render.ProcessResponse, error) {
	if req.Card == nil {
		return &render.ProcessResponse{}, fmt.Errorf("no card provided")
	}

	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth < 40 {
		termWidth = 80 // Safe fallback width
	}

	targetGridWidth := (termWidth * 3) / 4
	if targetGridWidth < 60 {
		targetGridWidth = 60 // Enforce a sane minimum size
	}

	return &render.ProcessResponse{
		FormattedFront: generateSplitView(req.Card, false, targetGridWidth),
		FormattedBack:  generateSplitView(req.Card, true, targetGridWidth),
	}, nil
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

	frontHeader := fmt.Sprintf("FRONT (QUESTION)")
	backHeader := fmt.Sprintf("BACK (ANSWER)")
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
