package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/retry"
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
	StartSend(ctx context.Context)
}

type Agent struct {
	Config     Configer
	httpClient *AgentHTTPClient
}

type AgentHTTPClient struct {
	httpClient *http.Client
	retry      retry.Retry
}

func (c *AgentHTTPClient) Do(ctx context.Context, r *http.Request) {
	action := func() (*http.Response, error) {
		return c.httpClient.Do(r)
	}
	f := c.retry.RetryAgent(ctx, action)
	f()
}

func (a *Agent) StartSend(ctx context.Context) {
	pollInterval := a.Config.GetPoolInterval()
	reportInterval := a.Config.GetReportInterval()
	addr := fmt.Sprintf("http://%s", a.Config.GetAddress())
	var metrics collector.Metric
	metricCollector := &collector.MetricCollector{}
	pollTimer := time.NewTicker(time.Duration(pollInterval * int64(time.Second)))
	reportTimer := time.NewTicker(time.Second * time.Duration(reportInterval))
	for {

		select {
		case <-pollTimer.C:
			logger.Log.Info("Poll metric")
			metrics = metricCollector.GetRunTimeMetrics()

		case <-reportTimer.C:
			logger.Log.Info("Send metric")
			sendBulkMetric(ctx, metrics, addr, a.httpClient)
		case <-ctx.Done():
			return
		}
	}
}

func CreateAgent() MetricSender {
	a := &Agent{}
	a.Config = config.GetConfig()
	a.httpClient = &AgentHTTPClient{httpClient: &http.Client{Timeout: time.Second * 5}, retry: retry.NewRetry(1, 2, 3)}
	return a
}

func sendBulkMetric(ctx context.Context, c collector.Metric, addr string, client *AgentHTTPClient) {
	if len(c.GaugeMetrics) == 0 && len(c.CounterMetrics) == 0 {
		return
	}
	d := make([]data.Metric, 0, len(c.GaugeMetrics)+len(c.CounterMetrics))
	for _, v := range c.GaugeMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeGauge, Value: &v.Value})
	}
	for _, v := range c.CounterMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeCounter, Delta: &v.Value})
	}
	sendMetricJSON(ctx, client, addr, d...)

}

func sendMetrics(ctx context.Context, c collector.Metric, addr string, client *AgentHTTPClient) {
	for _, m := range c.GaugeMetrics {
		sendMetricJSON(ctx, client, addr, data.Metric{ID: string(m.Name), MType: gauge, Value: &m.Value})
	}
	for _, m := range c.CounterMetrics {
		sendMetricJSON(ctx, client, addr, data.Metric{ID: string(m.Name), MType: counter, Delta: &m.Value})
	}
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

func sendMetricJSON(ctx context.Context, client *AgentHTTPClient, addr string, metric ...data.Metric) {
	data, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}
	method := "update"
	if len(metric) > 1 {
		method = "updates"
	}

	url := fmt.Sprintf("%s/%s", addr, method)
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(data)

	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	gzipWriter.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	client.Do(ctx, req)
}
