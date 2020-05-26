package app

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-mixins/log"
	"github.com/go-mixins/microservice/config"
	mw "github.com/go-mixins/microservice/http"
)

// App binds various parts together
type App struct {
	Config          *config.Config
	Logger          log.ContextLogger
	Handler         http.Handler
	wg              sync.WaitGroup
	stopChan        chan struct{}
	metricsHandler  http.Handler
	readinessChecks []mw.Check
	flushers        []interface{ Flush() }
	once            sync.Once
}

// FlushLogs pending hooks
func (app *App) FlushLogs() {
	for _, f := range app.flushers {
		f.Flush()
	}
}

// Run the app
func (app *App) Run() error {
	app.once.Do(func() {
		app.stopChan = make(chan struct{})
	})
	if err := app.connectLogs(); err != nil {
		app.Logger.Warnf("log hooks are not available: %v", err)
	}
	if err := app.connectTracing(); err != nil {
		app.Logger.Warnf("tracing is not available: %v", err)
	}
	if err := app.connectMetrics(); err != nil {
		app.Logger.Warnf("metrics export is not available: %v", err)
	}
	if connector, ok := app.Handler.(interface{ Connect() error }); ok {
		if err := connector.Connect(); err != nil {
			return err
		}
	}
	if closer, ok := app.Handler.(interface{ Close() error }); ok {
		defer closer.Close()
	}
	httpErrors, err := app.connectHTTP()
	if err != nil {
		return err
	}
	grpcErrors, err := app.connectGRPC()
	if err != nil {
		return err
	}
	app.Logger.Debugf("running in Timezone %v", time.Local)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	defer signal.Stop(interrupt)
	select {
	case sig := <-interrupt:
		close(app.stopChan)
		app.Logger.Infof("received %v", sig)
	case err = <-httpErrors:
		close(app.stopChan)
		app.Logger.Errorf("in HTTP handler: %+v", err)
	case err = <-grpcErrors:
		app.Logger.Errorf("in gRPC handler: %+v", err)
		close(app.stopChan)
	case <-app.stopChan:
		app.Logger.Info("force stop")
	}
	app.wg.Wait()
	return err
}

func (app *App) Stop() error {
	close(app.stopChan)
	app.wg.Wait()
	return nil
}
