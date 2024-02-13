package app

import (
	"context"

	"github.com/go-mixins/microservice/census/prom"
	"github.com/go-mixins/microservice/config"
	"github.com/go-mixins/microservice/opentracing"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/otel"
)

func (app *App) connectTracing() error {
	if app.TracerProvider == nil {
		var cfg opentracing.Config
		if err := config.Load(&cfg); err != nil {
			return err
		}
		exp, err := opentracing.New(context.Background(), cfg)
		if err != nil {
			return err
		}
		app.TracerProvider = exp
	}
	otel.SetTracerProvider(app.TracerProvider)
	return nil
}

func (app *App) connectMetrics() error {
	if app.MetricsExporter == nil {
		pe, err := prom.New(app.Config)
		if err != nil {
			return err
		}
		app.metricsHandler = pe
		app.MetricsExporter = pe
	}
	view.RegisterExporter(app.MetricsExporter)
	return nil
}
