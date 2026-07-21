// Package shared contains shared data between the host and plugins.
package shared

import (
	"context"

	render "github.com/dmclink/flash-cli/gen/go/render/v1"
	review "github.com/dmclink/flash-cli/gen/go/review/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

const (
	CAPABILITY_REVIEW_PROCESSOR = "review.v1.ReviewProcessorService"
	CAPABILITY_RENDER           = "render.v1.RenderService"
	CAPABILITY_ADD_CARD         = "addcard.v1.AddCardService"
	// TODO: add new capability keys here
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "FLASHCARD_CLI_PLUGIN_HANDSHAKE",
	MagicCookieValue: "flashcards-grpc-ecosystem-auth",
}

type RenderPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl render.RenderServiceServer
}

func (p *RenderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	render.RegisterRenderServiceServer(s, p.Impl)
	return nil
}

func (p *RenderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return render.NewRenderServiceClient(c), nil
}

type ReviewPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl review.ReviewProcessorServiceServer
}

func (p *ReviewPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	review.RegisterReviewProcessorServiceServer(s, p.Impl)
	return nil
}

func (p *ReviewPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return review.NewReviewProcessorServiceClient(c), nil
}

var PluginMap = map[string]plugin.Plugin{
	CAPABILITY_REVIEW_PROCESSOR: &ReviewPlugin{},
	CAPABILITY_RENDER:           &RenderPlugin{},
}
