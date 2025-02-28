package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
)

func main() {
	pollInterval := 2
	reportInterval := 10
	var metrics collector.Metric
	metricCollector := &collector.MetricCollector{}
	count := 0
	for {
		if count%pollInterval == 0 {
			metrics = metricCollector.GetRunTimeMetrics()
		}
		if count%reportInterval == 0 {
			for _, m := range metrics.GaugeMetrics {
				sendGaugeMetrics(string(m.Name), m.Value)
			}
			for _, m := range metrics.CounterMetrics {
				sendCounterMetrics(string(m.Name), int(m.Value))
			}
		}
		time.Sleep(time.Second)
		count++
	}
}

func sendGaugeMetrics(metricName string, value float64) {
	url := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%f", metricName, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-type", "text-plain")
	req.Header.Set("Content-Length", "0")
	if err != nil {
		fmt.Println("Error of creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error to sending request:", err)
		return
	}

	defer resp.Body.Close()
}

func sendCounterMetrics(metricName string, value int) {
	url := fmt.Sprintf("http://localhost:8080/update/counter/%s/%d", metricName, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-type", "text-plain")
	req.Header.Set("Content-Length", "0")
	if err != nil {
		fmt.Println("Error of creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error to sending request:", err)
		return
	}

	defer resp.Body.Close()
}
