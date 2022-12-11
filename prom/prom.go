package prom

import (
	"fmt"
	"strings"

	"github.com/go-mixins/microservice/v2/config"
	"go.opencensus.io/stats/view"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/google/wire"
)

var Set = wire.NewSet(New, wire.Bind(new(view.Exporter), new(*prometheus.Exporter)))

func New(cfg config.Config) (*prometheus.Exporter, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: strings.ReplaceAll(cfg.ServiceName, "-", "_"),
		ConstLabels: map[string]string{
			"environment": cfg.Environment,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}
	return pe, nil
}
