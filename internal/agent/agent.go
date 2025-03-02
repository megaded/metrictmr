package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
)

type Configer interface {
	GetAddress() string
	GetReportInterval() int64
	GetPoolInterval() int64
}

type MetricSender interface {
	StarSend()
}

type Agent struct {
	Config Configer
}

func (a *Agent) StarSend() {
	pollInterval := a.Config.GetPoolInterval()
	reportInterval := a.Config.GetReportInterval()
	addr := fmt.Sprintf("http://%s", a.Config.GetAddress())
	var metrics collector.Metric
	metricCollector := &collector.MetricCollector{}
	client := &http.Client{}
	var count int64 = 0
	for {
		if (count % pollInterval) == 0 {
			metrics = metricCollector.GetRunTimeMetrics()
		}
		if count%reportInterval == 0 {
			for _, m := range metrics.GaugeMetrics {
				sendGaugeMetrics(client, addr, string(m.Name), m.Value)
			}
			for _, m := range metrics.CounterMetrics {
				sendCounterMetrics(client, addr, string(m.Name), int(m.Value))
			}
		}
		time.Sleep(time.Second)
		count++
	}
}

func CreateAgent() MetricSender {
	a := &Agent{}
	a.Config = config.GetConfig()
	return a
}

func sendGaugeMetrics(client *http.Client, addr string, metricName string, value float64) {
	url := fmt.Sprintf("%s/update/gauge/%s/%f", addr, metricName, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-type", "text-plain")
	req.Header.Set("Content-Length", "0")
	if err != nil {
		fmt.Println("Error of creating request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error to sending request:", err)
		return
	}

	resp.Body.Close()
}

func sendCounterMetrics(client *http.Client, addr string, metricName string, value int) {
	url := fmt.Sprintf("%s/update/counter/%s/%d", addr, metricName, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-type", "text-plain")
	req.Header.Set("Content-Length", "0")
	if err != nil {
		fmt.Println("Error of creating request:", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error to sending request:", err)
		return
	}

	resp.Body.Close()
}
