package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var FlowGauge = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "warden_agent_flow_total",
		Help: "Number of network flows per agent session, verdict, and destination",
	},
	[]string{"session", "verdict", "destination"},
)
