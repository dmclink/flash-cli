package ext

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	review "github.com/dmclink/flash-cli/gen/go/review/v1"
	"github.com/dmclink/flash-cli/internal/app"
	"github.com/dmclink/flash-cli/internal/config"
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

// returned cleanup func must have its call deferred
func DispenseRenderer(a *app.App, name string) (Renderer, func(), error) {
	noOpCleanup := func() {}

	lookupString := a.Config.Resolve(config.KeyDefaultReviewRenderer, name, BASIC_RENDERER)

	switch lookupString {
	case BASIC_RENDERER:
		return renderer.BasicRenderer{}, noOpCleanup, nil
	}

	return dispensePlugin[render.RenderServiceClient, Renderer](
		a,
		lookupString,
		shared.CAPABILITY_RENDER,
		func(c render.RenderServiceClient) Renderer { return &rendererHostAdapter{client: c} },
	)
}

type ReviewProcessor interface {
	Process(ctx context.Context, dbCards []database.Flashcard, modifiers []string) ([]database.Flashcard, error)
}

func DispenseReviewProcessor(a *app.App, mode string) (ReviewProcessor, func(), error) {
	noOpCleanup := func() {}
	lookupString := a.Config.Resolve(config.KeyDefaultReviewMode, mode, SHUFFLE_MODE_KEY)

	// look for native review processor modes
	switch lookupString {
	case SHUFFLE_MODE_KEY:
		return reviewer.ShuffleMode{}, noOpCleanup, nil
	case LAST_REVIEW_MODE_KEY:
		return reviewer.LastReviewMode{}, noOpCleanup, nil
	case CREATED_AT_MODE_KEY:
		return reviewer.CreatedAtMode{}, noOpCleanup, nil
	case LINEAR_MODE_KEY:
		return reviewer.LinearMode{}, noOpCleanup, nil
	}

	return dispensePlugin(
		a,
		lookupString,
		shared.CAPABILITY_REVIEW_PROCESSOR,
		func(c review.ReviewProcessorServiceClient) ReviewProcessor {
			return &reviewProcessorHostAdapter{client: c}
		},
	)
}

func dispensePlugin[ClientType shared.Shutdownable, AdapterType any](a *app.App, pluginName string, capabilityName string, wrapClientFunc func(ClientType) AdapterType) (AdapterType, func(), error) {
	var zero AdapterType

	// findFunc returns PluginManifest but currently not being used, may be useful context for error handling for plugin info
	// to implement later
	_, binaryPath, err := FindPlugin(a, pluginName, capabilityName)
	if err != nil {
		return zero, nil, fmt.Errorf("scanning plugin directory | %w", err)
	}

	client := createClient(a, binaryPath)
	clientProtocol, err := client.Client()
	if err != nil {
		return zero, nil, fmt.Errorf("launching plugin process | %w", err)
	}

	rawClient, err := clientProtocol.Dispense(capabilityName)
	if err != nil {
		return zero, nil, fmt.Errorf("plugin does not support the render interface | %w", err)
	}

	grpcClient, ok := rawClient.(ClientType)
	if !ok {
		return zero, nil, fmt.Errorf("plugin type assertion failure: expected %T, got %T", (*ClientType)(nil), rawClient)
	}

	cleanup := func() {
		_, _ = grpcClient.Shutdown(context.Background(), &common.ShutdownRequest{})
		time.Sleep(15 * time.Millisecond)
		client.Kill()
	}

	return wrapClientFunc(grpcClient), cleanup, nil
}

func createClient(a *app.App, binaryPath string) *plugin.Client {
	cmd := exec.Command(binaryPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           a.Logger,
	})
}
