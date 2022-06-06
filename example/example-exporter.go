package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/prezhdarov/prometheus-exporter/example/api"

	exampleCollectors "github.com/prezhdarov/prometheus-exporter/example/collectors"

	"github.com/prezhdarov/prometheus-exporter/config"
	"github.com/prezhdarov/prometheus-exporter/exporter"

	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
)

const (
	exporterName = "Example Exporter"
	namespace    = "example"
)

var (
	listenAddress          = flag.String("http.address", ":9169", "Address and port to listen for http connections")
	maxRequests            = flag.Int("prom.maxRequests", 20, "Maximum number of parallel scrape requests. Use 0 to disable.")
	disableExporterTarget  = flag.Bool("disable.exporter.target", false, "Disable default target for /metrics path.")
	disableExporterMetrics = flag.Bool("disable.exporter.metrics", true, "Disable exporter metrics in /metrics path. Always enabled if /metrics target disabled")

	logLevel  = flag.String("log.level", "debug", "Log Level minimums. Available options are: debug,info,warn and error")
	logFormat = flag.String("log.format", "logfmt", "Log output format. Available options are: logfmt and json")
)

func setLogger(lf, ll *string) *promlog.Config {
	promlogFormat := &promlog.AllowedFormat{}
	promlogFormat.Set(*lf)

	promlogLevel := &promlog.AllowedLevel{}
	promlogLevel.Set(*ll)

	promlogConfig := &promlog.Config{}
	promlogConfig.Format = promlogFormat
	promlogConfig.Level = promlogLevel

	return promlogConfig
}

func usage() {
	const s = `
example-exporter collects metrics data from a fictional API. 
`
	config.Usage(s)
}

func main() {

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	config.Parse()

	logger := promlog.New(setLogger(logFormat, logLevel))

	level.Debug(logger).Log("disable exporter target is", disableExporterTarget)

	api.Load(logger)

	exampleCollectors.Load(logger)

	http.Handle("/metrics", exporter.CreateHandler(!*disableExporterMetrics, *disableExporterTarget, *maxRequests, namespace, logger))
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		exporter.CreateHandleFunc(w, r, namespace, "gateways", logger)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<head><title>Example Exporter</title></head>
			<body>
			<h1>Example Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			<p><a href="/probe">Probe</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "listening on", "address", listenAddress)

	server := &http.Server{Addr: *listenAddress}

	if err := web.ListenAndServe(server, "", logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

}
