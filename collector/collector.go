package collector

import (
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
)

type ClientAPI interface {
	Login(target string) error
	Logout() error
	Get(params map[string]string) (*[]byte, error)
	GetTarget() string
}

type ScrapeMetrics struct {
	Success  *prometheus.Desc
	Duration *prometheus.Desc
}

/*type MetricPackage struct {
	metricName        string
	metricDescription string
	labels            map[string]string
	metricValue       float64
}*/

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric, namespace string, client ClientAPI, params map[string]string) error
}

type CollectorSet struct {
	Collectors    map[string]Collector
	client        ClientAPI
	target        string
	namespace     string
	extraParams   map[string]string
	logger        log.Logger
	ScrapeMetrics ScrapeMetrics
}

const (
	DefaultEnabled  = true
	DefaultDisabled = false
)

var (
	registeredClientAPI    ClientAPI
	factories              = make(map[string]func(logger log.Logger) (Collector, error))
	collectorState         = make(map[string]*bool)
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
)

func RegisterAPI(clientAPI ClientAPI) {
	registeredClientAPI = clientAPI
}

func RegisterCollector(collector string, flag bool, factory func(logger log.Logger) (Collector, error)) {
	collectorState[collector] = &flag
	factories[collector] = factory
}

func NewCollectorSet(namespace, target string, params map[string]string, logger log.Logger) (*CollectorSet, error) {

	var sm ScrapeMetrics

	sm.Duration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)

	sm.Success = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)

	f := make(map[string]bool)
	collectors := make(map[string]Collector)

	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()

	for key, enabled := range collectorState {

		if !*enabled || (len(f) > 0 && !f[key]) {

			continue

		}

		if collector, ok := initiatedCollectors[key]; ok {

			collectors[key] = collector

		} else {

			collector, err := factories[key](log.With(logger, "collector", key))

			if err != nil {

				return nil, err

			}

			collectors[key] = collector

			initiatedCollectors[key] = collector

		}

	}

	return &CollectorSet{
		Collectors:    collectors,
		client:        registeredClientAPI,
		target:        target,
		namespace:     namespace,
		extraParams:   params,
		logger:        logger,
		ScrapeMetrics: sm,
	}, nil
}

func (cs *CollectorSet) Describe(ch chan<- *prometheus.Desc) {
	ch <- cs.ScrapeMetrics.Duration
	ch <- cs.ScrapeMetrics.Success
}

func (cs *CollectorSet) Collect(ch chan<- prometheus.Metric) {

	begin := time.Now()

	if err := cs.client.Login(cs.target); err != nil {

		level.Error(cs.logger).Log("msg", "Login failed", "target", cs.client.GetTarget(), "err", err)
		return

	} else {

		level.Debug(cs.logger).Log("msg", "Login successful", "target", cs.client.GetTarget())

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

	if err := cs.client.Logout(); err != nil {

		level.Error(cs.logger).Log("msg", "Logout failed", "target", cs.client.GetTarget(), "err", err)

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
