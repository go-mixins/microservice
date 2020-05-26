package grpc

import (
	mdGRPC "github.com/go-mixins/metadata/grpc"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

func ClientMiddleware(extraMW ...grpc.UnaryClientInterceptor) []grpc.DialOption {
	res := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithUnaryInterceptor(
			mdGRPC.UnaryClientInterceptor(),
		),
	}
	for _, e := range extraMW {
		res = append(res, grpc.WithUnaryInterceptor(e))
	}
	return res
}
