package exporter

import (
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &Collector{}

// Collector defines the structure for exporting latency metrics
type Collector struct {
  latencyMetric *prometheus.GaugeVec
  getMetrics    func() ([]NodeLatency, error) // Changed to return an array of NodeLatency
}

var LatestValues []NodeLatency
// NewCollector creates a new Collector instance
func NewCollector(getMetrics func() ([]NodeLatency, error)) *Collector {
  return &Collector{
    latencyMetric: prometheus.NewGaugeVec(
      prometheus.GaugeOpts{
        Name: "network_latency",
        Help: "Network latency in milliseconds for each node",
      },
      []string{"destination","timestamp"}, // Label for destination node name and timestamp 
    ),
    getMetrics: getMetrics,
  }
}

// Describe sends the descriptor of the metric to the channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
  c.latencyMetric.Describe(ch)
}

// Collect exports the latency metrics to the channel
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
  c.latencyMetric.Collect(ch)
}
func (c *Collector) Update() {
// collect latencies
  nodeLatencies, err := c.getMetrics()
  fmt.Println(nodeLatencies)
  LatestValues = nodeLatencies
  if err != nil {
    log.Printf("Error retrieving node latencies: %v", err)
    return
  }
  // export collected latencies
  for _, latency := range nodeLatencies {
    c.latencyMetric.WithLabelValues(latency.Destination,latency.Timestamp).Set(latency.Latency)
  }
}