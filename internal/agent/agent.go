package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/agent/data"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

type Configer interface {
	GetAddress() string
	GetReportInterval() int64
	GetPoolInterval() int64
}

type MetricSender interface {
	StartSend()
}

type Agent struct {
	Config Configer
}

func (a *Agent) StartSend() {
	pollInterval := a.Config.GetPoolInterval()
	reportInterval := a.Config.GetReportInterval()
	addr := fmt.Sprintf("http://%s", a.Config.GetAddress())
	var metrics collector.Metric
	metricCollector := &collector.MetricCollector{}
	client := &http.Client{Timeout: time.Second * 5}
	var count int64 = 0
	for {
		if (count % pollInterval) == 0 {
			metrics = metricCollector.GetRunTimeMetrics()
		}
		if count%reportInterval == 0 {
			for _, m := range metrics.GaugeMetrics {
				sendMetricJSON(client, addr, data.Metrics{ID: string(m.Name), MType: gauge, Value: &m.Value})
			}
			for _, m := range metrics.CounterMetrics {
				sendMetricJSON(client, addr, data.Metrics{ID: string(m.Name), MType: gauge, Delta: &m.Value})
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

func sendMetric(client *http.Client, addr string, metricType string, metricName collector.MetricName, value string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", addr, metricType, metricName, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-type", "text-plain")
	req.Header.Set("Content-Length", "0")
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
}

func sendMetricJSON(client *http.Client, addr string, metric data.Metrics) {
	data, err := json.Marshal(metric)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%s/update", addr)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-type", "application/json")
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
}
