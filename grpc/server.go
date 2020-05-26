package grpc

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-mixins/log"
	mdGRPC "github.com/go-mixins/metadata/grpc"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/go-mixins/microservice/json"
	grpcMW "github.com/grpc-ecosystem/go-grpc-middleware"
)

var grpcOnce sync.Once

// Replaceable functions
var (
	NowFunc = time.Now
)

// ServerMiddleware создает рекомендованный набор опций сервера
func ServerMiddleware(logger log.ContextLogger, extraMW ...grpc.UnaryServerInterceptor) []grpc.ServerOption {
	grpcOnce.Do(func() {
		if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
			logger.Errorf("registering gRPC views: %+v", err)
		}
	})
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.UnaryInterceptor(grpcMW.ChainUnaryServer(
			append([]grpc.UnaryServerInterceptor{
				RequestLogging(logger),
				mdGRPC.UnaryServerInterceptor(),
				ErrorsToStatus(),
			}, extraMW...)...,
		)),
	}
}

// RequestDebug включает логирование всех вызовов
func RequestDebug() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, rErr error) {
		reqPb, _ := req.(proto.Message)
		jd, _ := json.Encode(reqPb)
		log.Get(ctx).Debugf("received request: %s", jd)
		return handler(ctx, req)
	}
}

// RequestLogging инжектирует лог в контекст и ведет логи вызовов методов
func RequestLogging(logger log.ContextLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, rErr error) {
		logger := logger.WithContext(log.M{
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

// ErrorsToStatus пытается преобразовать ошибки обработчиков с status.Status
func ErrorsToStatus() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, rErr error) {
		res, err := handler(ctx, req)
		for {
			if _, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
				return res, err
			}
			if x, ok := err.(interface{ Unwrap() error }); ok {
				err = x.Unwrap()
				continue
			}
			if x, ok := err.(interface{ Cause() error }); ok {
				err = x.Cause()
				continue
			}
			return res, err
		}
	}
}
