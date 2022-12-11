package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/go-mixins/log"
	mdGRPC "github.com/go-mixins/metadata/grpc"
	"github.com/google/wire"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	grpcMW "github.com/grpc-ecosystem/go-grpc-middleware"
)

// Replaceable functions
var (
	NowFunc        = time.Now
	ServerSet      = wire.NewSet(ServerMiddleware)
	ServerSetDebug = wire.NewSet(ServerMiddlewareWithDebug)
)

func ServerMiddleware(logger log.ContextLogger) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.UnaryInterceptor(grpcMW.ChainUnaryServer(
			RequestLogging(logger),
			mdGRPC.UnaryServerInterceptor(),
		)),
	}
}

func ServerMiddlewareWithDebug(logger log.ContextLogger) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.UnaryInterceptor(grpcMW.ChainUnaryServer(
			RequestLogging(logger),
			RequestDebug(),
			mdGRPC.UnaryServerInterceptor(),
		)),
	}
}

// RequestDebug включает логирование всех вызовов
func RequestDebug() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, rErr error) {
		reqPb, _ := req.(proto.Message)
		jd, _ := protojson.Marshal(reqPb)
		log.Get(ctx).Debugf("received request: %s", jd)
		return handler(ctx, req)
	}
}

// RequestLogging инжектирует лог в контекст и ведет логи вызовов методов
func RequestLogging(logger log.ContextLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, rErr error) {
		logger = logger.WithContext(log.M{
			"method":   info.FullMethod,
			"trace_id": trace.FromContext(ctx).SpanContext().TraceID.String(),
		})
		ctx = log.With(ctx, logger)
		ts := NowFunc()
		defer func() {
			entry := log.M{
				"result": "success",
			}
			if err := recover(); err != nil {
				entry["result"] = "panic"
				logger.Errorf("panic: %+v", err)
				logger.Debugf("panic trace: %s", debug.Stack())
				defer panic(err)
			} else if rErr != nil {
				entry["result"] = "error"
				status, ok := status.FromError(rErr)
				if !ok {
					logger.Errorf("error: %+v", rErr)
				}
				entry["code"] = status.Code()
			}
			logger.WithContext(entry).Debugf("finished request in %v", NowFunc().Sub(ts))
		}()
		return handler(ctx, req)
	}
}
