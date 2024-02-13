package grpc

import (
	mdGRPC "github.com/go-mixins/metadata/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func ClientMiddleware(extraMW ...grpc.UnaryClientInterceptor) []grpc.DialOption {
	res := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(
			mdGRPC.UnaryClientInterceptor(),
		),
	}
	for _, e := range extraMW {
		res = append(res, grpc.WithUnaryInterceptor(e))
	}
	return res
}
