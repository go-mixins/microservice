package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"
)

// Config for app
type Config struct {
	Environment string        `envconfig:"ENVIRONMENT" default:"master"`
	ServiceName string        `envconfig:"SERVICE_NAME" required:"true"`
	HTTPPort    int           `envconfig:"HTTP_PORT" default:"5000"`
	HTTPPrefix  string        `envconfig:"HTTP_PREFIX"`
	GRPCPort    int           `envconfig:"GRPC_PORT" default:"8080"`
	Debug       bool          `envconfig:"DEBUG" default:"true"`
	SentryDSN   string        `envconfig:"SENTRY_DSN"`
	GraylogURI  string        `envconfig:"GRAYLOG_URI"`
	StopTimeout time.Duration `envconfig:"STOP_TIMEOUT"`
	Jaeger      struct {
		SamplingProbability float64 `envconfig:"TRACE_SAMPLING" default:"0.001"`
		JaegerAgentHost     string  `envconfig:"JAEGER_AGENT_HOST" default:"localhost"`
		JaegerAgentPort     int     `envconfig:"JAEGER_AGENT_PORT" default:"6831"`
	}
}

// Load parses env into configuration struct
func Load(dest interface{}) error {
	if err := envconfig.Process("", dest); err != nil {
		return xerrors.Errorf("parse config: %w", err)
	}
	return nil
}
