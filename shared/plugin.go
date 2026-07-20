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

type GenericPluginHandler[Req any, Res any] interface {
	Process(ctx context.Context, req Req) (Res, error)
}

type GenericPluginServer[Req any, Res any] struct {
	Handler func(context.Context, Req) (Res, error)
}

func (s *GenericPluginServer[Req, Res]) Process(ctx context.Context, req Req) (Res, error) {
	return s.Handler(ctx, req)
}

type GenericPluginClient[Req any, Res any] struct {
	CallFunc func(ctx context.Context, req Req, opts ...grpc.CallOption) (Res, error)
}

func (c *GenericPluginClient[Req, Res]) Process(ctx context.Context, req Req) (Res, error) {
	return c.CallFunc(ctx, req)
}

type GenericGRPCPlugin[ServerType any, ClientType any] struct {
	// required to add fallbacks and prevent net/rpc
	plugin.NetRPCUnsupportedPlugin
	RegisterServerFunc func(grpc.ServiceRegistrar, ServerType)
	NewClientFunc      func(grpc.ClientConnInterface) ClientType
	Impl               ServerType
}

func (p *GenericGRPCPlugin[S, C]) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	p.RegisterServerFunc(s, p.Impl)
	return nil
}

func (p *GenericGRPCPlugin[S, C]) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return p.NewClientFunc(c), nil
}

// PluginMap contains dispensible plugins.
var PluginMap = map[string]plugin.Plugin{
	CAPABILITY_REVIEW_PROCESSOR: &GenericGRPCPlugin[
		GenericPluginHandler[*review.ProcessRequest, *review.ProcessResponse],
		GenericPluginHandler[*review.ProcessRequest, *review.ProcessResponse]]{
		RegisterServerFunc: func(s grpc.ServiceRegistrar, impl GenericPluginHandler[*review.ProcessRequest, *review.ProcessResponse]) {
			review.RegisterReviewProcessorServiceServer(s,
				&GenericPluginServer[*review.ProcessRequest, *review.ProcessResponse]{
					Handler: impl.Process,
				})
		},
		NewClientFunc: func(cc grpc.ClientConnInterface) GenericPluginHandler[*review.ProcessRequest, *review.ProcessResponse] {
			pbClient := review.NewReviewProcessorServiceClient(cc)
			return &GenericPluginClient[*review.ProcessRequest, *review.ProcessResponse]{CallFunc: pbClient.Process}
		},
	},
	CAPABILITY_RENDER: &GenericGRPCPlugin[
		GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse],
		GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]]{
		RegisterServerFunc: func(s grpc.ServiceRegistrar, impl GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse]) {
			render.RegisterRenderServiceServer(s,
				&GenericPluginServer[*render.ProcessRequest, *render.ProcessResponse]{
					Handler: impl.Process,
				})
		},
		NewClientFunc: func(cc grpc.ClientConnInterface) GenericPluginHandler[*render.ProcessRequest, *render.ProcessResponse] {
			pbClient := render.NewRenderServiceClient(cc)
			return &GenericPluginClient[*render.ProcessRequest, *render.ProcessResponse]{CallFunc: pbClient.Process}
		},
	},

	// TODO: add new capabilities here
}
