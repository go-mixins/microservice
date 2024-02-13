package opentracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Config stores Jaeger-specific parameters
type Config struct {
	ServiceName string `envconfig:"SERVICE_NAME" default:"server"`
	Enabled     bool   `envconfig:"TRACING_ENABLED" default:"false"`
}

func New(ctx context.Context, cfg Config) (trace.TracerProvider, error) {
	if !cfg.Enabled {
		return noop.NewTracerProvider(), nil
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(cfg.ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("trace resource creation: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("gRPC trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(
			sdktrace.AlwaysSample(),
		),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res),
	)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
