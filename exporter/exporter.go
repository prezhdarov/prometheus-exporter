package exporter

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prezhdarov/prometheus-exporter/collector"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type eHandler struct {
	eHandler                http.Handler
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	disableExporterTarget   bool
	maxRequests             int
	logger                  *slog.Logger
}

func (h *eHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.eHandler.ServeHTTP(w, r)

}

func (h *eHandler) New(namespace, target string, params map[string]string) (http.Handler, error) {

	if h.disableExporterTarget {
		h.logger.Info("msg", "/metrics target is disabled. Serving exporter metrics only", nil)
		return promhttp.Handler(), nil
	}

	cl, err := collector.NewCollectorSet(namespace, target, params, h.logger)
	if err != nil {
		return nil, fmt.Errorf("could not create %s collector: %s", namespace, err)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(versioncollector.NewCollector(fmt.Sprintf("%s_exporter", namespace)))
	if err := registry.Register(&cl); err != nil {
		return nil, fmt.Errorf("could not register %s collector: %s", namespace, err)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, registry},
		promhttp.HandlerOpts{
			ErrorLog:      slog.NewLogLogger(h.logger.Handler(), slog.LevelError),
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      h.exporterMetricsRegistry,
		},
	)

	if h.includeExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(h.exporterMetricsRegistry, handler)
	}

	return handler, nil
}
