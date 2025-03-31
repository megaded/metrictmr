package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
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
				sendMetricJSON(client, addr, data.Metric{ID: string(m.Name), MType: gauge, Value: &m.Value})
			}
			for _, m := range metrics.CounterMetrics {
				sendMetricJSON(client, addr, data.Metric{ID: string(m.Name), MType: counter, Delta: &m.Value})
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

func sendMetricJSON(client *http.Client, addr string, metric data.Metric) {
	data, err := json.Marshal(metric)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%s/update", addr)
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(data)

	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	gzipWriter.Close()
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}

	defer resp.Body.Close()
}
