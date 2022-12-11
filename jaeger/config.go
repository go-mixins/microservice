package jaeger

import (
	"fmt"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/google/wire"
	"go.opencensus.io/trace"
)

var Set = wire.NewSet(New, wire.Bind(new(trace.Exporter), new(*jaeger.Exporter)))

// Config stores Jaeger-specific parameters
type Config struct {
	SamplingProbability float64 `envconfig:"TRACE_SAMPLING" default:"0.001"`
	JaegerAgentURI      string  `envconfig:"JAEGER_AGENT_URI"`
	ServiceName         string  `envconfig:"SERVICE_NAME" default:"server"`
}

func New(cfg Config) (*jaeger.Exporter, error) {
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: cfg.JaegerAgentURI,
		Process: jaeger.Process{
			ServiceName: cfg.ServiceName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create jaeger exporter: %+v", err)
	}
	return je, nil
}
