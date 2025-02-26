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
			fmt.Println("Получение метрики", time.Now())
			metrics = metricCollector.GetRunTimeMetrics()
		}
		if count%reportInterval == 0 {
			for _, m := range metrics.GaugeMetrics {
				url := fmt.Sprintf("/update/gauge/%s/%f", m.Name, m.Value)
				fmt.Println(url)
				resp, err := http.Post("http://localhost:8080"+url, "Content-Type: text/plain", nil)
				if err != nil {
					fmt.Println(err)
				}
				if resp.StatusCode != http.StatusOK {
					fmt.Println("Ошибка ", url)
				}
				fmt.Println("Успех ", url)
				resp.Body.Close()
			}
		}
		time.Sleep(time.Second)
		count++
	}
}
