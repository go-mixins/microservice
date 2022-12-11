package config

import (
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/kelseyhightower/envconfig"
)

// Config for app
type Config struct {
	Environment string        `envconfig:"ENVIRONMENT" default:"master"`
	ServiceName string        `envconfig:"SERVICE_NAME" required:"true"`
	HTTPPort    int           `envconfig:"HTTP_PORT" default:"8080"`
	HTTPPrefix  string        `envconfig:"HTTP_PREFIX"`
	GRPCPort    int           `envconfig:"GRPC_PORT"`
	Debug       bool          `envconfig:"DEBUG" default:"true"`
	SentryDSN   string        `envconfig:"SENTRY_DSN"`
	GraylogURI  string        `envconfig:"GRAYLOG_URI"`
	StopTimeout time.Duration `envconfig:"STOP_TIMEOUT"`
}

var Set = wire.NewSet(Load)

// Load parses env into configuration struct
func Load() (Config, error) {
	var dest Config
	if err := envconfig.Process("", &dest); err != nil {
		return dest, fmt.Errorf("parse config: %w", err)
	}
	return dest, nil
}
