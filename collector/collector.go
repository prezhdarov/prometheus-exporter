package collector

import (
	"sync"

	"github.com/go-kit/log"

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
