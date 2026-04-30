package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (cs *CollectorSet) Collect(ch chan<- prometheus.Metric) {

	begin := time.Now()

	clientData, err := cs.clientAPI.Login(cs.target, cs.logger)
	if err != nil {

		cs.logger.Error("Login failed", "target", clientData["target"], "err", err)
		return

	} else {

		cs.logger.Debug("Login successful", "target", clientData["target"])

	}

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(begin).Seconds(), "login") //Not really a collector, but helps get overall timing better

	wg := sync.WaitGroup{}

	cs.logger.Debug("number of collectors to scrape", "count", len(cs.Collectors))

	wg.Add(len(cs.Collectors))
	for name, c := range cs.Collectors {
		go func(name string, c Collector) {
			defer wg.Done()

			begin := time.Now()

			err := c.Update(ch, cs.namespace, cs.clientAPI, clientData, cs.extraParams)

			duration := time.Since(begin)

			var success float64

			if err != nil {
				cs.logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
				success = 0
			} else {
				cs.logger.Debug("collector scraped successfully", "target", clientData["target"].(string), "name", name, "duration_seconds", duration.Seconds())
				success = 1
			}
			ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Success, prometheus.GaugeValue, success, name)
		}(name, c)
	}

	wg.Wait()

	lobegin := time.Now()

	if err := cs.clientAPI.Logout(clientData, cs.logger); err != nil {

		cs.logger.Error("Logout failed", "target", clientData["target"], "err", err)

	} else {

		cs.logger.Debug("Logout successful", "target", clientData["target"])

	}

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(lobegin).Seconds(), "logout") //Same as Login above

	ch <- prometheus.MustNewConstMetric(cs.ScrapeMetrics.Duration, prometheus.GaugeValue, time.Since(begin).Seconds(), "all_collectors")
}
