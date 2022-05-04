package exporter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

//var extraParams = flag.String("extra.parameters", "", "Additional /probe parameters to pass to collector modules")

func CreateHandleFunc(w http.ResponseWriter, r *http.Request, namespace, extraParams string, logger log.Logger) {

	p := r.URL.Query()

	target := p.Get("target")

	level.Debug(logger).Log("msg", fmt.Sprintf("url: %s", r.URL.RawPath))

	params := make(map[string]string)

	for _, param := range strings.Split(extraParams, ",") {
		params[param] = p.Get(param)
	}

	if target == "" {
		level.Warn(logger).Log("msg", "No target specified.")
		return
	} else {
		level.Debug(logger).Log("msg", fmt.Sprintf("Scraping %s", params))
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
