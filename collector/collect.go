package collector

import (
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func (cs *CollectorSet) Collect(ch chan<- prometheus.Metric) {

	begin := time.Now()

	client := cs.client

	if err := client.Login(cs.target); err != nil {

		level.Error(cs.logger).Log("msg", "Login failed", "target", client.GetTarget(), "err", err)
		return

	} else {

		level.Debug(cs.logger).Log("msg", "Login successful", "target", client.GetTarget())

	}

	wg := sync.WaitGroup{}

	level.Debug(cs.logger).Log("msg", "number of collectors to scrape", len(cs.Collectors))

	wg.Add(len(cs.Collectors))
	for name, c := range cs.Collectors {
		go func(name string, c Collector) {
			run(name, cs.namespace, c, ch, &cs.extraParams, cs.client, &cs.ScrapeMetrics, cs.logger)
			wg.Done()
		}(name, c)
	}

	wg.Wait()

	if err := client.Logout(); err != nil {

		level.Error(cs.logger).Log("msg", "Logout failed", "target", client.GetTarget(), "err", err)

	}

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(begin).Seconds(), "all_collectors")
}

func run(name, namespace string, c Collector, ch chan<- prometheus.Metric, params *map[string]string, client ClientAPI, sm *ScrapeMetrics, logger log.Logger) {
	begin := time.Now()

	err := c.Update(ch, namespace, client, *params)

	duration := time.Since(begin)

	var success float64

	if err != nil {
		level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		success = 0
	} else {
		level.Debug(logger).Log("msg", "collector scraped successfuly", "name", name, "duration", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(sm.Duration, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(sm.Success, prometheus.GaugeValue, success, name)

}
