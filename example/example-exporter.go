package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prezhdarov/prometheus-exporter/example/api"

	exampleCollectors "github.com/prezhdarov/prometheus-exporter/example/collectors"

	"github.com/prezhdarov/prometheus-exporter/config"
	"github.com/prezhdarov/prometheus-exporter/exporter"

	"github.com/prometheus/common/promslog"
	"github.com/prometheus/exporter-toolkit/web"
)

const (
	// This namespace is what each and every metric provided by this exporter will begin with. Choose wisely as this is what you will get famous with!
	namespace = "example"
)

var (
	// The address and port to bind to. Address can be omitted. not sure if I have maxRequests implemented, probably standard for node_exporter functions.
	listenAddress = flag.String("http.address", ":9169", "Address and port to listen for http connections")
	maxRequests   = flag.Int("prom.maxRequests", 20, "Maximum number of parallel scrape requests. Use 0 to disable.")
	// Wether to disable exporter own metrics and disable default /metrics target (only /probe is usable). Note that if both disabled /metrics will return exporter metrics regardless.
	disableExporterTarget  = flag.Bool("disable.exporter.target", false, "Disable default target for /metrics path.")
	disableExporterMetrics = flag.Bool("disable.exporter.metrics", true, "Disable exporter metrics in /metrics path. Always enabled if /metrics target disabled")

	// These two are used for promlog to configure the logger. Quite self-explanatory
	logLevel  = flag.String("log.level", "debug", "Log Level minimums. Available options are: debug,info,warn and error")
	logFormat = flag.String("log.format", "logfmt", "Log output format. Available options are: logfmt and json")
)

// This is where the magic of -h/-help is happening. I know - not so much of a magic in this case.
func usage() {
	const s = `
	example-exporter collects metrics data from a fictional API. 
	`
	config.Usage(s)
}

func main() {

	// Here we set all output to stdout (both error and standard) display a simple usage message and parse all command-line flags
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	config.Parse()

	// This exporter uses prometheys logger (promlog) extensively and even passes it to the collectors to log their endeavours
	logger := promslog.New(config.SetLogger(logFormat, logLevel))

	// Tells us if we have the /metrics is disabled in debug in case we begin to wonder why nothing comes out..
	logger.Debug("disable exporter target is", fmt.Sprintf("%v", disableExporterTarget), nil)

	// This is my awkward way of loading the so called API reader. Don't judge!
	api.Load(logger)

	//  Enable all configured collectors. Note each collector can be enabled or disabled by default and its state can be altered using a flag. Also feels a bit lame..
	exampleCollectors.Load(logger)

	// Here prometheus dudes create a handler for /metrics with its own registry, which can be reused again and again with every scrape.
	// Also the /probe handle function is created with its own separate registry for each target and after the scrape is destroyed... I think...
	http.Handle("/metrics", exporter.CreateHandler(!*disableExporterMetrics, *disableExporterTarget, *maxRequests, namespace, logger))
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		exporter.CreateHandleFunc(w, r, namespace, "gateways", logger)
	})
	// A simple description should someone get lost and end in the exporter root :)
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

	// Tell us essentially if exporter has managed to bind to the configured address (or shout out an error if not)
	logger.Info("msg", "listening on", "address", *listenAddress, nil)

	// Prometheus toolkit changed a little since version 0.8 (I think) and now configuring the http server takes a slightly more complicated approach that requires a Web Config.
	// Below is the final step needed to start the exporter - create a http server... I think again.. It is borrowed.
	server := &http.Server{}

	if err := web.ListenAndServe(server, config.WebConfig(listenAddress), logger); err != nil {
		logger.Error("err", fmt.Sprint(err), nil)
		os.Exit(1)
	}

}
