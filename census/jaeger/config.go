package jaeger

import (
	"fmt"

	"contrib.go.opencensus.io/exporter/jaeger"
)

// Config stores Jaeger-specific parameters
type Config struct {
	SamplingProbability float64 `envconfig:"TRACE_SAMPLING" default:"0.001"`
	JaegerAgentHost     string  `envconfig:"JAEGER_AGENT_HOST" default:"localhost"`
	JaegerAgentPort     int     `envconfig:"JAEGER_AGENT_PORT" default:"6831"`
	ServiceName         string  `envconfig:"SERVICE_NAME" default:"server"`
}

func New(cfg Config) (*jaeger.Exporter, error) {
	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: fmt.Sprintf("%s:%d", cfg.JaegerAgentHost, cfg.JaegerAgentPort),
		Process: jaeger.Process{
			ServiceName: cfg.ServiceName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create jaeger exporter: %+v", err)
	}
	return je, nil
}
