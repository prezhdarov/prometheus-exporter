package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func (cs *CollectorSet) Collect(ch chan<- prometheus.Metric) {

	begin := time.Now()

	clientData, err := cs.clientAPI.Login(cs.target)
	if err != nil {

		level.Error(cs.logger).Log("msg", "Login failed", "target", clientData["target"], "err", err)
		return

	} else {

		level.Debug(cs.logger).Log("msg", "Login successful", "target", clientData["target"])

	}

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(begin).Seconds(), "login") //Not really a collector, but helps get overall timing better

	wg := sync.WaitGroup{}

	level.Debug(cs.logger).Log("msg", fmt.Sprintf("number of collectors to scrape: %d", len(cs.Collectors)))

	wg.Add(len(cs.Collectors))
	for name, c := range cs.Collectors {
		go func(name string, c Collector) {

			begin := time.Now()

			err := c.Update(ch, cs.namespace, cs.clientAPI, clientData, cs.extraParams)

			duration := time.Since(begin)

			var success float64

			if err != nil {
				level.Error(cs.logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
				success = 0
			} else {
				level.Debug(cs.logger).Log("msg", "collector scraped successfully", "target", clientData["target"].(string), "name", name, "duration", duration.Seconds())
				success = 1
			}
			ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Success, prometheus.GaugeValue, success, name)

			wg.Done()
		}(name, c)
	}

	wg.Wait()

	lobegin := time.Now()

	if err := cs.clientAPI.Logout(clientData); err != nil {

		level.Error(cs.logger).Log("msg", "Logout failed", "target", clientData["target"], "err", err)

	} else {

		level.Debug(cs.logger).Log("msg", "Logout successful", "target", clientData["target"])

	}

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(lobegin).Seconds(), "logout") //Same as Login above

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(begin).Seconds(), "all_collectors")
}
