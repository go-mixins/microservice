package grpc

import (
	mdGRPC "github.com/go-mixins/metadata/grpc"
	"github.com/google/wire"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

var ClientSet = wire.NewSet(ClientMiddleware)

func ClientMiddleware() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithUnaryInterceptor(mdGRPC.UnaryClientInterceptor()),
	}
}
