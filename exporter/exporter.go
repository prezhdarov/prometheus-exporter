package exporter

import (
	"fmt"
	stdlog "log"
	"net/http"

	"forti-exporter/libs/collector"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

type eHandler struct {
	eHandler                http.Handler
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	disableExporterTarget   bool
	maxRequests             int
	logger                  log.Logger
}

func (h *eHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.eHandler.ServeHTTP(w, r)

}

func (h *eHandler) New(namespace, target string, params map[string]string) (http.Handler, error) {

	if h.disableExporterTarget {
		level.Info(h.logger).Log("msg", "/metrics target is disabled. Serving exporter metrics only")
		return promhttp.Handler(), nil
	}

	cl, err := collector.NewCollectorSet(namespace, target, params, h.logger)
	if err != nil {
		return nil, fmt.Errorf("could not create %s collector: %s", namespace, err)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(version.NewCollector(fmt.Sprintf("%s_exporter", namespace)))
	if err := registry.Register(&cl); err != nil {
		return nil, fmt.Errorf("could not register %s collector: %s", namespace, err)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, registry},
		promhttp.HandlerOpts{
			ErrorLog:      stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      h.exporterMetricsRegistry,
		},
	)

	if h.includeExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(h.exporterMetricsRegistry, handler)
	}

	return handler, nil
}
