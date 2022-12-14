package app

import (
	"github.com/FiboChain/fbc/x/common/monitor"
)

var (
	// init monitor prometheus metrics
	orderMetrics  = monitor.DefaultOrderMetrics(monitor.DefaultPrometheusConfig())
	streamMetrics = monitor.DefaultStreamMetrics(monitor.DefaultPrometheusConfig())
)
