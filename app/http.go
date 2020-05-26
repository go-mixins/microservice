package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-mixins/log"

	mw "github.com/go-mixins/microservice/http"
)

func (app *App) stopTimeout() time.Duration {
	if app.Config.StopTimeout != 0 {
		return app.Config.StopTimeout
	}
	return time.Second * 5
}

func (app *App) connectHTTP() (<-chan error, error) {
	errorChan := make(chan error, 1)
	handler := mw.WithHealth(app.Handler, app.readinessChecks...)
	handler = mw.WithMetrics(handler, app.metricsHandler)
	handler = mw.WithLog(handler, app.Logger.WithContext(log.M{"logger": "http"}))
	handler = mw.WithTracing(handler)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.HTTPPort),
		Handler: handler,
	}
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer app.Logger.Infof("stopped HTTP server")
		go func() {
			errorChan <- server.ListenAndServe()
		}()
		select {
		case <-app.stopChan:
			ctx, cancel := context.WithTimeout(context.Background(), app.stopTimeout())
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				app.Logger.Warnf("stopping HTTP server: %+v", err)
			}
		}
	}()
	return errorChan, nil
}
