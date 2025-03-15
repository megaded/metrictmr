package agent

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
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
				sendMetric(client, addr, gauge, m.Name, strconv.FormatFloat(m.Value, 'f', 6, 64))
			}
			for _, m := range metrics.CounterMetrics {
				sendMetric(client, addr, counter, m.Name, strconv.Itoa(int(m.Value)))
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
