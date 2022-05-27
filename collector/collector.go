package collector

import (
	"flag"
	"fmt"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
)

type ClientAPI interface {
	Login(target string) (map[string]interface{}, error)
	Logout(loginData map[string]interface{}) error
	Get(loginData, extraConfig map[string]interface{}) (interface{}, error)
}

type ScrapeMetrics struct {
	Success  *prometheus.Desc
	Duration *prometheus.Desc
}

// Collector is the interface a collector has to implement.
type Collector interface {
	Update(ch chan<- prometheus.Metric, namespace string, clientAPI ClientAPI, clientData map[string]interface{}, extraParams map[string]string) error
}

type CollectorSet struct {
	Collectors    map[string]Collector
	clientAPI     ClientAPI
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

var disableDefaultCollector = flag.Bool("disable.default.collectors", DefaultDisabled, "If set only explicitly enabled collectors will be enabled")

var (
	registeredClientAPI    ClientAPI
	factories              = make(map[string]func(logger log.Logger) (Collector, error))
	collectorState         = make(map[string]*bool)
	initiatedCollectorsMtx = sync.Mutex{}
	initiatedCollectors    = make(map[string]Collector)
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == fmt.Sprintf("collector.%s", name) {
			found = true
		}
	})
	return found
}

func disableDefaultCollectors() {
	for collector := range collectorState {
		if !isFlagPassed(collector) {
			*collectorState[collector] = false
		}
	}
}

func RegisterAPI(clientAPI ClientAPI) {
	registeredClientAPI = clientAPI
}

func RegisterCollector(collector string, flag *bool, factory func(logger log.Logger) (Collector, error)) {

	collectorState[collector] = flag
	factories[collector] = factory
}

func NewCollectorSet(namespace, target string, params map[string]string, logger log.Logger) (CollectorSet, error) {

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

	if *disableDefaultCollector {
		disableDefaultCollectors()
	}

	for key, enabled := range collectorState {

		if !*enabled || (len(f) > 0 && !f[key]) {
			level.Debug(logger).Log("msg", fmt.Sprintf("Collector %s is disabled", key))
			continue
		}

		level.Debug(logger).Log("msg", fmt.Sprintf("Collector %s is enabled", key))

		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](log.With(logger, "collector", key))
			if err != nil {
				return CollectorSet{}, err
			}

			collectors[key] = collector

			initiatedCollectors[key] = collector

		}

	}

	return CollectorSet{
		Collectors:    collectors,
		clientAPI:     registeredClientAPI,
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
