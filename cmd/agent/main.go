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
	host := "http://localhost:8080"
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
				url := fmt.Sprintf("/update/counter/%s/%d", m.Name, m.Value)
				fmt.Println(host + url)
				resp, err := http.Post(host+url, "text/plain", http.NoBody)
				if err != nil {
					fmt.Println(err)
				}
				resp.Body.Close()
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
