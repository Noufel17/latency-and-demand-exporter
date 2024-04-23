package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)


type NodeLatency struct {
  Destination string  `json:"destination"`
  Latency    float64 `json:"latency"`
}

func GetNodeLatencies() ([]NodeLatency, error) {
  var nodeLatencies []NodeLatency

  // Get a Kubernetes clientset
  clientset, err := getClientset()
  if err != nil {
    return nil, err
  }

  // List all nodes in the cluster
  nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
  fmt.Println("Kubernetes nodes")
  fmt.Println(nodes)
  if err != nil {
    return nil, err
  }

  for index, node := range nodes.Items {
    fmt.Println("node data")
    fmt.Println(node)
    // mesure latency to worker nodes only
    if node.Labels["node-role.kubernetes.io/worker"] == "true" {
      nodeName := node.Name; // should work
      destIP := node.Status.Addresses[0].Address // should give the reachable IP between all nodes

      // Use iperf3 to measure latency
      latency, err := measureLatency(destIP)
      fmt.Printf("single latency %d \n",index)
      fmt.Println(latency)
      if err != nil {
        log.Printf("Error measuring latency to node %s : addresse %s : %v",nodeName ,destIP, err)
        continue
      }

      nodeLatencies = append(nodeLatencies, NodeLatency{
        Destination: nodeName,
        Latency:    latency,
      })
    }
  }

  return nodeLatencies, nil
}

func measureLatency(destIP string) (float64, error) {
  // Define iperf3 execution options (adjust as needed)
  cmd := exec.Command("iperf3", "-c", destIP, "-p", "5201", "-t", "1", "-f", "json")

  // Execute iperf3 and capture output
  output, err := cmd.CombinedOutput()
  if err != nil {
    return 0, fmt.Errorf("error running iperf3: %w", err)
  }

  // Parse iperf3 JSON output to extract average latency
  avgLatency, err := parseIperf3Latency(output)
  if err != nil {
    return 0, fmt.Errorf("error parsing iperf3 output: %w", err)
  }

  return avgLatency, nil
}

func parseIperf3Latency(output []byte) (float64, error) {
  var data map[string]interface{}
  err := json.Unmarshal(output, &data)
  if err != nil {
    return 0, fmt.Errorf("error unmarshalling iperf3 JSON: %w", err)
  }
  fmt.Println("structured output data");
  fmt.Println(data);

  // Access nested data based on iperf3 JSON output format (adjust as needed)
  intervalsInterface, ok := data["intervals"]
  if !ok {
    return 0, errors.New("missing 'intervals' field in iperf3 JSON")
  }

  intervals, ok := intervalsInterface.([]interface{})
  if !ok {
    return 0, errors.New("invalid format for 'intervals' field in iperf3 JSON")
  }

  if len(intervals) == 0 {
    return 0, errors.New("no intervals found in iperf3 JSON")
  }

  // Assuming the first interval contains average latency (check iperf3 documentation)
  firstInterval := intervals[0].(map[string]interface{})
  avgLatencyInterface, ok := firstInterval["avg_rtt_ms"]
  if !ok {
    return 0, errors.New("missing 'avg_rtt_ms' field in iperf3 JSON")
  }

  avgLatency, ok := avgLatencyInterface.(float64)
  if !ok {
    return 0, errors.New("invalid format for 'avg_rtt_ms' field in iperf3 JSON")
  }

  return avgLatency, nil
}


func getClientset() (*kubernetes.Clientset, error) {
  // Create a config object using in-cluster config
  config, err := rest.InClusterConfig()
  if err != nil {
    return nil, err
  }

  // Create a new clientset from the config
  return kubernetes.NewForConfig(config)
}