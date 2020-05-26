package app

import (
	"fmt"
	"net"

	"github.com/go-mixins/log"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	gRPCmw "github.com/go-mixins/microservice/grpc"
)

func (app *App) connectGRPC() (<-chan error, error) {
	errorChan := make(chan error, 1)
	grpcConnector, ok := app.Handler.(interface{ ConnectGRPC(*grpc.Server) error })
	if !ok {
		return nil, nil
	}
	var mw []grpc.UnaryServerInterceptor
	if optsProvider, ok := app.Handler.(interface {
		GRPCInterceptors() []grpc.UnaryServerInterceptor
	}); ok {
		mw = append(mw, optsProvider.GRPCInterceptors()...)
	}
	opts := gRPCmw.ServerMiddleware(app.Logger.WithContext(log.M{"logger": "gRPC"}), mw...)
	grpcServer := grpc.NewServer(opts...)
	if err := grpcConnector.ConnectGRPC(grpcServer); err != nil {
		return nil, err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.GRPCPort))
	if err != nil {
		return nil, xerrors.Errorf("opening listener: %w", err)
	}
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer app.Logger.Infof("stopped gRPC server")
		go func() {
			errorChan <- grpcServer.Serve(lis)
		}()
		select {
		case <-app.stopChan:
			grpcServer.GracefulStop()
		}
	}()
	return errorChan, nil
}
