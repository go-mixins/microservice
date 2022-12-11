package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-mixins/log"
	"github.com/go-mixins/microservice/v2/config"
	httpmw "github.com/go-mixins/microservice/v2/http"
	"github.com/google/wire"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"gocloud.dev/server"
	"gocloud.dev/server/driver"
	"gocloud.dev/server/health"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var Set = wire.NewSet(
	wire.Struct(new(App),
		"Config",
		"Logger",
		"Handler",
		"HealthChecks",
		"MetricsExporter",
		"TraceExporter",
		"DefaultSamplingPolicy",
		"Driver",
		"GRPCServer",
	),
	wire.Value(&server.DefaultDriver{}),
	wire.Bind(new(driver.Server), new(*server.DefaultDriver)),
)

type ServerStopTimeout time.Duration

// App binds various parts together
type App struct {
	Config                config.Config
	Logger                log.ContextLogger
	Handler               http.Handler
	HealthChecks          []health.Checker
	MetricsExporter       view.Exporter
	TraceExporter         trace.Exporter
	DefaultSamplingPolicy trace.Sampler
	Driver                driver.Server
	GRPCServer            *grpc.Server

	flushers []interface{ Flush() } `wire:"-"`
}

// FlushLogs pending hooks
func (app *App) FlushLogs() {
	for _, f := range app.flushers {
		f.Flush()
	}
}

// Run the app
func (app *App) Run() error {
	g, ctx := errgroup.WithContext(context.Background())
	if err := app.connectLogs(); err != nil {
		app.Logger.Warnf("log hooks are not available: %v", err)
	}
	handler := app.Handler
	if handler != nil {
		handler = httpmw.WithLog(handler, app.Logger)
	}
	if app.MetricsExporter != nil {
		view.RegisterExporter(app.MetricsExporter)
		ochttp.ServerLatencyView.TagKeys = append(ochttp.ServerLatencyView.TagKeys, ochttp.KeyServerRoute)
		if err := view.Register(ochttp.DefaultServerViews...); err != nil {
			return fmt.Errorf("registering HTTP views: %w", err)
		}
		if h, ok := app.MetricsExporter.(http.Handler); ok {
			handler = httpmw.WithMetrics(handler, h)
		}
	}
	server := server.New(handler, &server.Options{
		RequestLogger:         nil, // we have our own logger middleware
		HealthChecks:          app.HealthChecks,
		TraceExporter:         app.TraceExporter,
		DefaultSamplingPolicy: app.DefaultSamplingPolicy,
		Driver:                app.Driver,
	})
	g.Go(func() error {
		return server.ListenAndServe(fmt.Sprintf(":%d", app.Config.HTTPPort))
	})
	if grpcServer := app.GRPCServer; grpcServer != nil {
		if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
			return fmt.Errorf("registering gRPC views: %w", err)
		}
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", app.Config.GRPCPort))
		if err != nil {
			return fmt.Errorf("opening gRPC listener: %w", err)
		}
		g.Go(func() error {
			return grpcServer.Serve(lis)
		})
	}
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		errChan <- g.Wait()
	}()
	app.Logger.Debugf("running in Timezone %v", time.Local)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	defer signal.Stop(interrupt)
	select {
	case sig := <-interrupt:
		app.Logger.Infof("received %v", sig)
		if app.GRPCServer != nil {
			app.GRPCServer.GracefulStop()
		}
		ctx, cancel := context.WithTimeout(ctx, app.Config.StopTimeout)
		defer cancel()
		return server.Shutdown(ctx)
	case err := <-errChan:
		return err
	}
}
