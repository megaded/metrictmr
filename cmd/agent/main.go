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
				url := fmt.Sprintf("/update/gauge/%s/%f", m.Name, m.Value)
				fmt.Println(host + url)
				resp, err := http.Post(host+url, "text/plain", http.NoBody)
				if err != nil {
					fmt.Println(err)
				}
				resp.Body.Close()

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
