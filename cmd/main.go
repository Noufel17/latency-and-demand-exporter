package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"noufel/latency-node-exporter/cmd/exporter"
	"time"

	"math/rand"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {
  var (
		promPort = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
	)
	flag.Parse()

	// called on each collector.Collect which is called on each prometheus interval default 15 sec
	getMetrics := func() ([]exporter.NodeLatency,error){
		return exporter.GetNodeLatencies()
	}

	// make prometheus aware of our collectors
	collector := exporter.NewCollector(getMetrics)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	go func() {
		for {
			collector.Update()
			time.Sleep(time.Second * 10)
		}
	}()

  promMux := http.NewServeMux()
  http.HandleFunc("/export", returnLatestValue)
  http.HandleFunc("/demand", exportDemand)
  
  log.Printf("listening exporter on %q/export", ":9150")
  if err := http.ListenAndServe(":9150", nil); err != nil {
    log.Fatalf("cannot start exporter: %s", err)
  }
  promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
  promMux.Handle("/metrics", promHandler)

  port := fmt.Sprintf(":%d", *promPort)
  log.Printf("starting exporter on %q/metrics", port)
  if err := http.ListenAndServe(port, promMux); err != nil {
    log.Fatalf("cannot start exporter: %s", err)
  }
}

func returnLatestValue(w http.ResponseWriter, r *http.Request){
  defer r.Body.Close()
  
	jsonData, err := json.Marshal(exporter.LatestValues)
  if err != nil {
    // Handle marshalling error
    http.Error(w, "Error marshalling data to JSON: "+err.Error(), http.StatusInternalServerError)
    return
  }

  // Set the content type header to application/json
  w.Header().Set("Content-Type", "application/json")

  // Write the JSON data to the response body
  w.Write(jsonData)
}

type Demand struct {
	Demand float64
}

func roundFloat(val float64, precision uint) float64 {
    ratio := math.Pow(10, float64(precision))
    return math.Round(val*ratio) / ratio
}

func exportDemand(w http.ResponseWriter, r *http.Request){
  defer r.Body.Close()
	randomDemand := rand.Float64()

  
	jsonData, err := json.Marshal(Demand{Demand:roundFloat(randomDemand,5)})
  if err != nil {
    // Handle marshalling error
    http.Error(w, "Error marshalling data to JSON: "+err.Error(), http.StatusInternalServerError)
    return
  }

  // Set the content type header to application/json
  w.Header().Set("Content-Type", "application/json")

  // Write the JSON data to the response body
  w.Write(jsonData)
}