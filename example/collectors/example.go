package exampleCollectors

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prezhdarov/prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Here we define what will be appended to exporter namespace in each and every metric. This has to be unique for each collector in the set.
const (
	testSubsystem = "test"
)

var testCollectorFlag = flag.Bool("test.collector", collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", testSubsystem, collector.DefaultEnabled))

// The collector itself
type testCollector struct {
	logger log.Logger
}

// Load is my silly way to.... well, load the collector set - i.e. all enabled collectors in the package at exporter start.
func Load(logger log.Logger) {
	level.Info(logger).Log("msg", "Loading Example collector set")
}

// This adds the collector to the set of collectors to be used during Collect phase. The testCollectorFlag is used to either set the collector as enabled or disabled (bool)
func init() {
	collector.RegisterCollector("test", testCollectorFlag, NewTestCollector)
}

// The Baron survives another drowning... just..
func NewTestCollector(logger log.Logger) (collector.Collector, error) {
	return &testCollector{logger}, nil
}

// This is where the magic happens. Here clientAPI.Get can be consumed and metrics created with the result and pushed out.
func (c *testCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	extraConfig := make(map[string]interface{}, 0)

	clientAPI.Get(loginData, extraConfig, c.logger)

	// This is a simple metric of type Gauge (could be Counter for all it matters too).
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, testSubsystem, "some_fake_metric"),
			"This is a fake metric... but is it?", nil, nil,
		), prometheus.GaugeValue, 1.0,
	)

	// This is a simple metric, but with a timestamp.
	ch <- prometheus.NewMetricWithTimestamp(
		time.Now(), prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, testSubsystem, "some_fake_metric_with_time"),
				"This is also a fake metric... with a timestamp!", nil, nil,
			), prometheus.GaugeValue, 1.0,
		),
	)

	return nil
}
