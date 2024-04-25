package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	//"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
)


type NodeLatency struct {
  Destination string  `json:"destination"`
  Latency    float64 `json:"latency"`
  Timestamp string `json:"timestamp"`
}

type Iperf3Data struct {
  Start    struct {
    Timestamp struct {
      Timesecs int64 `json:"timesecs"`
    } `json:"timestamp"`
  } `json:"start"`
  End      struct {
    Streams []struct {
      Sender struct {
        MeanRtt float64 `json:"mean_rtt"`
      } `json:"sender"`
    } `json:"streams"`
  } `json:"end"`
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
  if err != nil {
    return nil, err
  }

  for _, node := range nodes.Items {
    // mesure latency to worker nodes only
    if node.Labels["upf-candidate"] == "true" {
      nodeName := node.Name; // should work
      destIP := node.Status.Addresses[0].Address // should give the reachable IP between all nodes

      // Use iperf3 to measure latency
      latency,timestamp, err := measureLatency(destIP)
      if err != nil {
        log.Printf("Error measuring latency to node %s : addresse %s : %v",nodeName ,destIP, err)
        continue
      }

      nodeLatencies = append(nodeLatencies, NodeLatency{
        Destination: nodeName,
        Latency:    latency,
        Timestamp: timestamp.Format("2006-01-02 15:04:05 UTC"),
      })
    }
  }

  return nodeLatencies, nil
}

func measureLatency(destIP string) (float64,time.Time, error) {
  // Define iperf3 execution options (adjust as needed)
  cmd := exec.Command("iperf3", "-c", destIP, "-p", "5201", "-t", "5", "-J")

  // Execute iperf3 and capture output
  output, err := cmd.CombinedOutput()
  if err != nil {
    return 0, metav1.Now().Time,fmt.Errorf("error running iperf3: %w", err)
  }

  // Parse iperf3 JSON output to extract average latency
  avgLatency,timestamp, err := parseIperf3Latency(output)
  if err != nil {
    return 0, metav1.Now().Time,fmt.Errorf("error parsing iperf3 output: %w", err)
  }

  return avgLatency,timestamp, nil
}



func parseIperf3Latency(output []byte) (float64, time.Time, error) {
  var data Iperf3Data
  err := json.Unmarshal(output, &data)
  if err != nil {
    return 0, time.Time{}, fmt.Errorf("error unmarshalling iperf3 JSON: %w", err)
  }

  if len(data.End.Streams) == 0 {
    // Handle empty streams (connection not established)
    return 0, time.Time{}, errors.New("connection not established")
  }

  // Extract mean_rtt
  meanRtt := data.End.Streams[0].Sender.MeanRtt

  // Extract and format timestamp (assuming timesecs is Unix timestamp)
  timestamp := time.Unix(data.Start.Timestamp.Timesecs, 0)

  return meanRtt, timestamp, nil
}


func getClientset() (*kubernetes.Clientset, error) {
  // Create a config object using in-cluster config
  config, err := rest.InClusterConfig()
  // kubeadm cluster token config file location
  //config, err := clientcmd.BuildConfigFromFlags("", "/etc/kubernetes/kubeadm-client.conf")
  if err != nil {
    return nil, err
  }

  // Create a new clientset from the config
  return kubernetes.NewForConfig(config)
}