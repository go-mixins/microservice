package prom

import (
	"fmt"
	"strings"

	"github.com/go-mixins/microservice/config"

	"contrib.go.opencensus.io/exporter/prometheus"
)

func New(cfg *config.Config) (*prometheus.Exporter, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: strings.Replace(cfg.ServiceName, "-", "_", -1),
		ConstLabels: map[string]string{
			"environment": cfg.Environment,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %+v", err)
	}
	return pe, nil
}
