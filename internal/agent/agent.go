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
	"time"

	"github.com/megaded/metrictmr/internal/agent/collector"
	"github.com/megaded/metrictmr/internal/agent/config"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/retry"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

func (c *AgentHTTPClient) Do(ctx context.Context, eg *errgroup.Group, r *http.Request) {
	action := func() (*http.Response, error) {
		return c.httpClient.Do(r)
	}
	f := c.retry.RetryAgent(ctx, action)
	eg.Go(f)
}

func (a *Agent) StartSend(ctx context.Context) {
	pollInterval := a.Config.GetPoolInterval()
	addr := fmt.Sprintf("http://%s", a.Config.GetAddress())
	key := a.Config.GetKey()
	rateLimit := a.Config.GetRateLimit()
	mch := make(chan collector.Metric, rateLimit)
	metricCollector := &collector.MetricCollector{}

	group, ctxCancel := errgroup.WithContext(ctx)

	sendMetric := func(ct context.Context, eg *errgroup.Group, mch chan collector.Metric) {
		for w := 0; w <= rateLimit; w++ {
			go worker(ctxCancel, eg, addr, key, a.httpClient, mch)
		}
	}

	collectMetrics := func(ct context.Context, mch chan collector.Metric) {
		for {
			select {
			case <-ct.Done():
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
	go sendMetric(ctxCancel, group, mch)
	if err := group.Wait(); err != nil {
		logger.Log.Error("Agent error", zap.Error(err))
	}
}

func CreateAgent() MetricSender {
	a := &Agent{}
	a.Config = config.GetConfig()
	a.httpClient = &AgentHTTPClient{httpClient: &http.Client{Timeout: time.Second * 5}, retry: retry.NewRetry(1, 2, 3)}
	return a
}

func worker(ctx context.Context, eg *errgroup.Group, addr string, key string, client *AgentHTTPClient, jobs <-chan collector.Metric) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				for m := range jobs {
					logger.Log.Info("Send metric")
					sendBulkMetric(ctx, eg, m, addr, key, client)
				}
			}
		}
	}
}

func sendBulkMetric(ctx context.Context, eg *errgroup.Group, c collector.Metric, addr string, key string, client *AgentHTTPClient) {
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
	sendMetricJSON(ctx, eg, client, addr, key, d...)
}

func sendMetricJSON(ctx context.Context, eg *errgroup.Group, client *AgentHTTPClient, addr string, key string, metric ...data.Metric) {
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
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	client.Do(ctx, eg, req)
}
