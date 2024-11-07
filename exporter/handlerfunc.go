package exporter

import (
	"fmt"
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
		logger.Warn("msg", "No target specified.", nil)
		return
	} else {
		logger.Debug("msg", fmt.Sprintf("Scraping %s", target), nil)
	}

	h := &eHandler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  false,
		disableExporterTarget:   false,
		maxRequests:             20,
		logger:                  logger,
	}

	if handler, err := h.New(namespace, target, params); err != nil {
		panic(fmt.Sprintf("could not create metrics handler: %s", err))
	} else {
		h.eHandler = handler

	}

	h.eHandler.ServeHTTP(w, r)
}
