package opentracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Config stores Jaeger-specific parameters
type Config struct {
	SamplingProbability float64 `envconfig:"TRACE_SAMPLING" default:"0.001"`
	ServiceName         string  `envconfig:"SERVICE_NAME" default:"server"`
	Enabled             bool    `envconfig:"TRACING_ENABLED" default:"true"`
}

func New(cfg Config) (trace.TracerProvider, error) {
	if !cfg.Enabled {
		return noop.NewTracerProvider(), nil
	}

	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
