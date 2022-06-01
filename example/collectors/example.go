package exampleCollectors

import (
	"flag"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prezhdarov/prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	testSubsystem = "test"
)

var testCollectorFlag = flag.Bool("test.collector", collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", testSubsystem, collector.DefaultEnabled))

type testCollector struct {
	logger log.Logger
}

func Load(logger log.Logger) {
	level.Info(logger).Log("msg", "Loading Example collector set")
}

func init() {
	collector.RegisterCollector("test", testCollectorFlag, NewTestCollector)
}

func NewTestCollector(logger log.Logger) (collector.Collector, error) {
	return &testCollector{logger}, nil
}

func (c *testCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, testSubsystem, "some_fake_metric"),
			"This is a fake metric... but is it?", nil, nil,
		), prometheus.GaugeValue, 1.0,
	)

	return nil
}
