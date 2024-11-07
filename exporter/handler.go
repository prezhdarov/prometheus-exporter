package exporter

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func CreateHandler(includeExporerMetrics, disableExporterTarget bool, maxRequests int, namespace string, logger *slog.Logger) *eHandler {
	h := &eHandler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporerMetrics,
		disableExporterTarget:   disableExporterTarget,
		maxRequests:             maxRequests,
		logger:                  logger,
	}

	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}

	if handler, err := h.New(namespace, "", map[string]string{}); err != nil {
		panic(fmt.Sprintf("could not create metrics handler: %s", err))
	} else {
		h.eHandler = handler

	}

	return h
}
