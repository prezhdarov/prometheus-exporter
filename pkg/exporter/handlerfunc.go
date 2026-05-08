package exporter

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// bufferedResponseWriter captures the handler response so the status code
// can be overridden before committing bytes to the real ResponseWriter.
type bufferedResponseWriter struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func (b *bufferedResponseWriter) Header() http.Header         { return b.header }
func (b *bufferedResponseWriter) Write(p []byte) (int, error) { return b.body.Write(p) }
func (b *bufferedResponseWriter) WriteHeader(statusCode int)  { b.statusCode = statusCode }

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

	buf := &bufferedResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
	h.eHandler.ServeHTTP(buf, r)

	for k, vs := range buf.header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if h.ScrapeFailed() {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(buf.statusCode)
	}
	w.Write(buf.body.Bytes()) //nolint:errcheck
}
