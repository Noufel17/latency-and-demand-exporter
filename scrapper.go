package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	DestIP := ""
	url := "http://+" + DestIP + ":9150/metrics"

	// Define the PromQL query to fetch the latest network latency
	query := `last_over_time(network_latency{destination="kind-worker"}[1m])`
	req, err := http.NewRequest("GET", url+"?query="+query, nil)
	if err != nil {
		fmt.Errorf("Error")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Errorf("Error")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Error")
	}

	// Parse the result to a float64
	result := strings.TrimSpace(string(body))
  fmt.Println("Result:"+result)

}