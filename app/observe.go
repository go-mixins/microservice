package app

import (
	"github.com/go-mixins/microservice/census/jaeger"
	"github.com/go-mixins/microservice/census/prom"
	"github.com/go-mixins/microservice/config"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func (app *App) connectTracing() error {
	if app.TraceExporter == nil {
		var cfg jaeger.Config
		if err := config.Load(&cfg); err != nil {
			return err
		}
		exp, err := jaeger.New(cfg)
		if err != nil {
			return err
		}
		app.TraceExporter = exp
		defer trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(cfg.SamplingProbability)})
	}
	trace.RegisterExporter(app.TraceExporter)
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
