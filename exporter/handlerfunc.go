package exporter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func CreateHandleFunc(w http.ResponseWriter, r *http.Request, namespace, extraParams string, logger log.Logger) {

	p := r.URL.Query()

	target := p.Get("target")

	params := make(map[string]string)

	for _, param := range strings.Split(extraParams, ",") {
		params[param] = p.Get(param)
	}

	if target == "" {
		level.Warn(logger).Log("msg", "No target specified.")
		return
	} else {
		level.Debug(logger).Log("msg", fmt.Sprintf("Scraping %s", target))
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
