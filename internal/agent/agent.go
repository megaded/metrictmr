package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/retry"
	"go.uber.org/zap"
)

const (
	gauge      = "gauge"
	counter    = "counter"
	hashHeader = "HashSHA256"
)

type Configer interface {
	GetAddress() string
	GetReportInterval() int64
	GetPoolInterval() int64
	GetKey() string
	GetRateLimit() int
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
	key        string
}

func (c *AgentHTTPClient) Do(ctx context.Context, ech chan error, r *http.Request) {
	action := func() (*http.Response, error) {
		return c.httpClient.Do(r)
	}
	f := c.retry.RetryAgent(ctx, action)
	err := f()
	if err != nil {
		go func() {
			defer close(ech)
			ech <- err
			logger.Log.Error("Ошибка при отправки метрик", zap.Error(err))
		}()
	}
}

func (a *Agent) StartSend(ctx context.Context) {
	ctxCancel, cancelFunc := context.WithCancel(ctx)
	pollInterval := a.Config.GetPoolInterval()
	reportInterval := a.Config.GetReportInterval()
	addr := fmt.Sprintf("http://%s", a.Config.GetAddress())
	key := a.Config.GetKey()
	rateLimit := a.Config.GetRateLimit()
	mch := make(chan collector.Metric, rateLimit)
	metricCollector := &collector.MetricCollector{}
	var wg sync.WaitGroup
	wg.Add(1)

	ech := make(chan error)
	sendMetric := func(ct context.Context, ec chan error, mch chan collector.Metric) {
		time.Sleep(time.Second * time.Duration(reportInterval))
		for w := 0; w <= rateLimit; w++ {
			go worker(ctxCancel, ech, reportInterval, addr, key, a.httpClient, mch)
		}
	}

	collectMetrics := func(ct context.Context, mch chan collector.Metric) {
		for {
			select {
			case <-ct.Done():
				wg.Done()
				close(mch)
				logger.Log.Info("Вышли Collect metric")
				return
			case <-time.After(time.Second * time.Duration(pollInterval)):
				logger.Log.Info("Collect metric")
				m := metricCollector.GetRunTimeMetrics()
				mch <- m
			}
		}
	}
	go collectMetrics(ctxCancel, mch)
	go sendMetric(ctxCancel, ech, mch)
	go func() {
		defer wg.Done()
		for e := range ech {
			logger.Log.Error("Agent error", zap.Error(e))
			cancelFunc()
			return
		}
	}()
	wg.Wait()
}

func CreateAgent() MetricSender {
	a := &Agent{}
	a.Config = config.GetConfig()
	a.httpClient = &AgentHTTPClient{httpClient: &http.Client{Timeout: time.Second * 5}, retry: retry.NewRetry(1, 2, 3)}
	return a
}

func worker(ctx context.Context, ec chan error, delay int64, addr string, key string, client *AgentHTTPClient, jobs <-chan collector.Metric) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				for m := range jobs {
					logger.Log.Info("Send metric")
					sendBulkMetric(ctx, ec, m, addr, key, client)
					time.Sleep(time.Second * time.Duration(delay))
				}
			}
		}
	}
}

func sendBulkMetric(ctx context.Context, ech chan error, c collector.Metric, addr string, key string, client *AgentHTTPClient) {
	if len(c.GaugeMetrics) == 0 && len(c.CounterMetrics) == 0 {
		logger.Log.Info("Отправка метрик. Метрик нет")
		return
	}
	d := make([]data.Metric, 0, len(c.GaugeMetrics)+len(c.CounterMetrics))
	for _, v := range c.GaugeMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeGauge, Value: &v.Value})
	}
	for _, v := range c.CounterMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeCounter, Delta: &v.Value})
	}
	sendMetricJSON(ctx, ech, client, addr, key, d...)
}

func sendMetrics(ctx context.Context, ech chan error, c collector.Metric, addr string, key string, client *AgentHTTPClient) {
	for _, m := range c.GaugeMetrics {
		sendMetricJSON(ctx, ech, client, addr, key, data.Metric{ID: string(m.Name), MType: gauge, Value: &m.Value})
	}
	for _, m := range c.CounterMetrics {
		sendMetricJSON(ctx, ech, client, addr, key, data.Metric{ID: string(m.Name), MType: counter, Delta: &m.Value})
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

func sendMetricJSON(ctx context.Context, ech chan error, client *AgentHTTPClient, addr string, key string, metric ...data.Metric) {
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
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(data)
		hash := hex.EncodeToString(h.Sum(nil))
		req.Header.Set(hashHeader, hash)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	client.Do(ctx, ech, req)
}
