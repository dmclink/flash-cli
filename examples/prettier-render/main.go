package main

import (
	"context"
	"fmt"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	"github.com/dmclink/flash-cli/shared"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
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
	grpcServer     *grpc.Server
}

// THIS EXACT SIGNATURE IS MANDATORY.
// Go's compiler forces this method name to be 'Process' (capital P)
// and the arguments must match your protoc-generated Go request models exactly.
func (h *RenderHandler) Process(ctx context.Context, req *render.ProcessRequest) (*render.ProcessResponse, error) {
	if req.Card == nil {
		return &render.ProcessResponse{}, fmt.Errorf("no card provided to plugin")
	}

	// Safely pass execution to your clean renderer.go module code
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

// Init is called once on plugin startup, but it is safe to completely omit this method
// as it errors out silently and falls back to sane defaults for these strings
// Otherwise, set these strings to add a welcome/startup banner and alter the instruction
// messages at the bottom of the render
func (h *RenderHandler) Init(ctx context.Context, req *render.InitRequest) (*render.InitResponse, error) {
	// req is empty and unused

	// return any desired metadata
	return &render.InitResponse{
		StartupBanner:    startupBanner,
		InstructionFront: instructionFront,
		InstructionBack:  instructionBack,
	}, nil
}

// Shutdown is called by the host program when rendering stops or receives a kill signal
func (h *RenderHandler) Shutdown(ctx context.Context, req *common.ShutdownRequest) (*common.ShutdownResponse, error) {
	// Plugins written in Go can just be a noop for shutdown as go-plugin pluging.Serve() handles kill signals gracefully
	return &common.ShutdownResponse{}, nil
}

// Initialization Entry Point: Negotiates connections with the Go core app host.
// To build your own layout engine, leave this file alone and edit renderer.go.
func main() {
	rendererLogic := &PrettierRenderer{}
	renderImpl := &RenderHandler{CustomRenderer: rendererLogic}

	// Just pass your handler straight to the concrete struct
	pluginMap := map[string]plugin.Plugin{
		shared.CAPABILITY_RENDER: &shared.RenderPlugin{Impl: renderImpl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         pluginMap,
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			s := grpc.NewServer(opts...)
			renderImpl.grpcServer = s
			return s
		},
	})
}
