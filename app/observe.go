package app

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/xerrors"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func (app *App) connectTracing() error {
	cfg := app.Config.Jaeger
	serviceName := os.Getenv("SERVICE_NAME")
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: fmt.Sprintf("%s:%d", cfg.JaegerAgentHost, cfg.JaegerAgentPort),
		Process: jaeger.Process{
			ServiceName: serviceName,
		},
	})
	if err != nil {
		return xerrors.Errorf("connect jaeger exporter: %w", err)
	}
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(cfg.SamplingProbability)})
	return nil
}

func (app *App) connectMetrics() error {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: strings.Replace(os.Getenv("SERVICE_NAME"), "-", "_", -1),
		ConstLabels: map[string]string{
			"environment": app.Config.Environment,
		},
	})
	if err != nil {
		return xerrors.Errorf("connect prometheus exporter: %w", err)
	}
	view.RegisterExporter(pe)
	app.metricsHandler = pe
	return nil
}
