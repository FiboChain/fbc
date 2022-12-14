package monitor

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"strconv"
	"sync"
)

var (
	portMetrics     *PortMetrics
	initPortMetrics sync.Once
)

// PortMetrics monitors connecting number
type PortMetrics struct {
	ConnectingNums metrics.Gauge
}

// GetPortMetrics returns Metrics build using Prometheus client library if Prometheus is enabled
// Otherwise, it returns no-op Metrics
func GetPortMetrics() *PortMetrics {
	initPortMetrics.Do(func() {
		if DefaultPrometheusConfig().Prometheus {
			portMetrics = NewPortMetrics()
		} else {
			portMetrics = NopPortMetrics()
		}
	})

	return portMetrics
}

// NewPortMetrics returns a pointer of a new PortMetrics object
func NewPortMetrics(labelsAndValues ...string) *PortMetrics {
	return &PortMetrics{
		ConnectingNums: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: xNameSpace,
			Subsystem: portSubSystem,
			Name:      "connecting",
			Help:      "connecting number of each port",
		}, []string{portSubSystem}).With(labelsAndValues...),
	}
}

// NopPortMetrics returns a pointer of a no-op Metrics
func NopPortMetrics() *PortMetrics {
	return &PortMetrics{
		ConnectingNums: discard.NewGauge(),
	}
}

//SetConnectingNums sets connectingNums by connecting number of each port
func (portMetrics *PortMetrics) SetConnectingNums(connectingMap map[uint64]int) {
	if nil == connectingMap {
		return
	}
	for port, num := range connectingMap {
		portMetrics.ConnectingNums.With(portSubSystem, strconv.FormatUint(port, 10)).Set(float64(num))
	}
}
