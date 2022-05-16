package svc

import (
	"google.golang.org/grpc"
)

// Specified GRPC Server Options for the server.
// Can only be used by the Init phase.
// Can be used, for example, to add interceptors by components.
type GRPCOptions struct{ opts []grpc.ServerOption }

func (g *GRPCOptions) Add(opts ...grpc.ServerOption) { g.opts = append(g.opts, opts...) }
