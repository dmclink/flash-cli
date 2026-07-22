package shared

import (
	"context"

	common "github.com/dmclink/flash-cli/gen/go/common/v1"
	"google.golang.org/grpc"
)

// Shutdownable defines a client plugin that can be shutdown. All plugins should match this interface
type Shutdownable interface {
	Shutdown(ctx context.Context, in *common.ShutdownRequest, opts ...grpc.CallOption) (*common.ShutdownResponse, error)
}
