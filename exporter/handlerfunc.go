package exporter

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func CreateHandleFunc(w http.ResponseWriter, r *http.Request, namespace, extraParams string, logger *slog.Logger) {

	p := r.URL.Query()

	target := p.Get("target")

	params := make(map[string]string)

	for _, param := range strings.Split(extraParams, ",") {
		params[param] = p.Get(param)
	}

	if target == "" {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
		return
	} else {
		logger.Debug("scraping target", "target", target)
	}

	h := &eHandler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  false,
		disableExporterTarget:   false,
		maxRequests:             20,
		logger:                  logger,
	}

	if handler, err := h.New(namespace, target, params); err != nil {
		http.Error(w, "could not create metrics handler: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		h.eHandler = handler

	}

	h.eHandler.ServeHTTP(w, r)
}
