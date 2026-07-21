package main

import (
	"context"
	"fmt"

	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	"github.com/dmclink/flash-cli/shared"
	"github.com/hashicorp/go-plugin"
)

// Network Router: Receives incoming network wire data from the host CLI app
// and safely routes it to your custom renderer.go logic.
//
// INTERFACE CONTRACT WARNING FOR DEVELOPERS:
// This struct satisfies your shared.GenericPluginHandler interface.
// Do NOT rename the 'Process' method or alter its input/output parameters.
// Doing so will prevent the compiler from registering your plugin.
type RenderHandler struct {
	// CustomRenderer exposes your layout implementation logic.
	CustomRenderer *PrettierRenderer
}

// THIS EXACT SIGNATURE IS MANDATORY.
// Go's compiler forces this method name to be 'Process' (capital P)
// and the arguments must match your protoc-generated Go request models exactly.
func (h *RenderHandler) Process(ctx context.Context, req *render.ProcessRequest) (*render.ProcessResponse, error) {
	if req.Card == nil {
		return &render.ProcessResponse{}, fmt.Errorf("no card provided to plugin")
	}

	// Safely pass execution to your clean renderer.py style module code
	front, back, progress := h.CustomRenderer.RenderCard(
		req.Card,
		req.CurrentCardNum,
		req.TotalCardCount,
		req.UnparsedModifiers,
	)

	return &render.ProcessResponse{
		FormattedFront: front,
		FormattedBack:  back,
		Progress:       progress,
	}, nil
}

// Initialization Entry Point: Negotiates connections with the Go core app host.
// To build your own layout engine, leave this file alone and edit renderer.go.
func main() {
	// Instantiate your unique styling layout configuration
	rendererLogic := &PrettierRenderer{}
	renderImpl := &RenderHandler{CustomRenderer: rendererLogic}

	// Register the interface hook mapping inside the plugin network topology
	pluginMap := map[string]plugin.Plugin{
		shared.CAPABILITY_RENDER: &shared.GenericGRPCPlugin[
			shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse],
			shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]]{
			Impl:               renderImpl,
			RegisterServerFunc: shared.PluginMap[shared.CAPABILITY_RENDER].(*shared.GenericGRPCPlugin[shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse], shared.GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]]).RegisterServerFunc,
		},
	}

	// Start the background gRPC daemon subprocess loop
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
