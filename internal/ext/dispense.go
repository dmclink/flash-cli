package ext

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	review "github.com/dmclink/flash-cli/gen/go/review/v1"
	"github.com/dmclink/flash-cli/internal/database"
	"github.com/dmclink/flash-cli/internal/renderer"
	reviewer "github.com/dmclink/flash-cli/internal/reviewer/modes"
	"github.com/dmclink/flash-cli/shared"
	"github.com/hashicorp/go-plugin"
)

const (
	SHUFFLE_MODE_KEY     = "shuffle"
	LAST_REVIEW_MODE_KEY = "lastreview"
	CREATED_AT_MODE_KEY  = "createdat"
	LINEAR_MODE_KEY      = "linear"
)

const (
	BASIC_RENDERER = "basic"
)

type Renderer interface {
	Render(ctx context.Context, card database.Flashcard, cardNum int, cardCount int, modifiers []string) (string, string, string, error)
	Init(ctx context.Context) (string, string, string, error)
}

func DispenseRenderer(name string) (Renderer, func(), error) {
	noOpCleanup := func() {}

	// TODO: when renderer defaults are setup in config, change this from BASIC to whatever they have setup
	// only fallback to basic if config not exists
	if name == "" || name == "default" {
		name = BASIC_RENDERER
	}

	switch name {
	case BASIC_RENDERER:
		return renderer.BasicRenderer{}, noOpCleanup, nil
	}

	manifest, binaryPath, err := FindRendererPlugin(name)
	if err != nil {
		return nil, nil, fmt.Errorf("scanning plugin directory | %w", err)
	}
	if manifest == nil {
		return nil, nil, fmt.Errorf("unknown renderer '%s' | no matching plugin manifest found", name)
	}

	// silentLogger := hclog.New(&hclog.LoggerOptions{
	// 	Name:   "discard",
	// 	Output: io.Discard,
	// 	Level:  hclog.Off,
	// })
	cmd := exec.Command(binaryPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		// Logger:           silentLogger,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("launching plugin process | %w", err)
	}

	raw, err := rpcClient.Dispense(shared.CAPABILITY_RENDER)
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin does not support the render interface | %w", err)
	}

	renderClient, ok := raw.(render.RenderServiceClient)
	if !ok {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin type assertion failure: expected render.RenderServiceClient")
	}

	renderPlugin := &rendererHostAdapter{client: renderClient}

	cleanup := func() {
		_, _ = renderClient.Shutdown(context.Background(), &render.ShutdownRequest{})
		time.Sleep(15 * time.Millisecond)
		client.Kill()
	}

	// need to defer the returned func here to cleanup process
	return renderPlugin, cleanup, nil
}

type ReviewProcessor interface {
	Process(ctx context.Context, dbCards []database.Flashcard, modifiers []string) ([]database.Flashcard, error)
}

func DispenseReviewProcessor(mode string) (ReviewProcessor, func(), error) {
	noOpCleanup := func() {}

	// TODO: when review mode defaults are setup in config, change this from SHUFFLE to whatever they have setup
	// only fallback to shuffle if not exists
	if mode == "" || mode == "default" {
		mode = SHUFFLE_MODE_KEY
	}

	// look for native review processor modes
	switch mode {
	case SHUFFLE_MODE_KEY:
		return reviewer.ShuffleMode{}, noOpCleanup, nil
	case LAST_REVIEW_MODE_KEY:
		return reviewer.LastReviewMode{}, noOpCleanup, nil
	case CREATED_AT_MODE_KEY:
		return reviewer.CreatedAtMode{}, noOpCleanup, nil
	case LINEAR_MODE_KEY:
		return reviewer.LinearMode{}, noOpCleanup, nil
	}

	manifest, binaryPath, err := FindReviewPlugin(mode)
	if err != nil {
		return nil, nil, fmt.Errorf("scanning plugin directory | %w", err)
	}
	if manifest == nil {
		return nil, nil, fmt.Errorf("unknown review mode '%s' | no matching plugin manifest found", mode)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("launching plugin process | %w", err)
	}

	raw, err := rpcClient.Dispense(shared.CAPABILITY_REVIEW_PROCESSOR)
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin does not support the review_processor interface | %w", err)
	}

	reviewClient, ok := raw.(review.ReviewProcessorServiceClient)
	if !ok {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin type assertion failure: unexpected underlying struct")
	}

	reviewerPlugin := &reviewProcessorHostAdapter{client: reviewClient}

	// need to defer the returned func here to cleanup process
	return reviewerPlugin, client.Kill, nil
}
