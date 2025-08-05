package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
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
	GetCryptoKeyPath() string
}

type Agent struct {
	Config     Configer
	httpClient *AgentHTTPClient
	Protocol   string
	hostIP     string
}

type AgentHTTPClient struct {
	httpClient *http.Client
	retry      retry.Retry
	key        string
}

func (c *AgentHTTPClient) Do(ctx context.Context, r *http.Request) error {
	action := func() (*http.Response, error) {
		return c.httpClient.Do(r)
	}
	return c.retry.RetryAgent(ctx, action)()
}

func (a *Agent) StartSend(ctx context.Context) {
	pollInterval := a.Config.GetPoolInterval()
	addr := fmt.Sprintf("%s://%s", a.Protocol, a.Config.GetAddress())
	key := a.Config.GetKey()
	rateLimit := a.Config.GetRateLimit()
	mch := make(chan collector.Metric, rateLimit)
	metricCollector := &collector.MetricCollector{}

	group, ctxCancel := errgroup.WithContext(ctx)

	for w := 0; w <= rateLimit; w++ {
		group.Go(func() error {
			return httpWorker(ctxCancel, addr, key, a.httpClient, a.hostIP, mch)
		})

	}
	group.Go(func() error {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()
		defer close(mch)
		for {
			select {
			case <-ctxCancel.Done():
				return ctxCancel.Err()
			case <-ticker.C:
				m := metricCollector.GetRunTimeMetrics()
				select {
				case mch <- m:
				case <-ctxCancel.Done():
					return ctxCancel.Err()
				}

			}
		}
	})

	if err := group.Wait(); err != nil {
		logger.Log.Error("Agent error", zap.Error(err))
	}
}

func CreateHttpClient() *Agent {
	a := &Agent{}
	a.Config = config.GetConfig()
	ip, err := getLocalIP()
	if err != nil {
		a.hostIP = ip.String()
	}
	a.httpClient = &AgentHTTPClient{httpClient: &http.Client{Timeout: time.Second * 5}, retry: retry.NewRetry(1, 2, 3)}
	protocol := "http"
	if a.Config.GetCryptoKeyPath() != "" {
		protocol = "https"
		tlsConfig, err := createTLSConfig()
		if err != nil {
			logger.Log.Error(err.Error())
			panic(err)
		}
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		a.httpClient.httpClient.Transport = transport
	}
	a.Protocol = protocol
	return a
}

func getLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP, nil
}

func createTLSConfig() (*tls.Config, error) {
	return &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}, nil
}

func httpWorker(ctx context.Context, addr string, key string, client *AgentHTTPClient, ip string, jobs <-chan collector.Metric) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-jobs:
			if !ok {
				return nil
			}
			if err := sendBulkMetric(ctx, m, addr, key, client, ip); err != nil {
				logger.Log.Warn("send metric error", zap.Error(err))
			}
		}
	}
}

func sendBulkMetric(ctx context.Context, c collector.Metric, addr string, key string, client *AgentHTTPClient, ip string) error {
	if len(c.GaugeMetrics) == 0 && len(c.CounterMetrics) == 0 {
		logger.Log.Info("Отправка метрик. Метрик нет")
		return nil
	}
	d := make([]data.Metric, 0, len(c.GaugeMetrics)+len(c.CounterMetrics))
	for _, v := range c.GaugeMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeGauge, Value: &v.Value})
	}
	for _, v := range c.CounterMetrics {
		d = append(d, data.Metric{ID: string(v.Name), MType: data.MTypeCounter, Delta: &v.Value})
	}
	return sendMetricJSON(ctx, client, addr, key, ip, d...)
}

func sendMetricJSON(ctx context.Context, client *AgentHTTPClient, addr string, key string, ip string, metric ...data.Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error(err.Error())
		return err
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
		return err
	}
	gzipWriter.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}
	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(data)
		hash := hex.EncodeToString(h.Sum(nil))
		req.Header.Set(hashHeader, hash)
	}

	if ip != "" {
		req.Header.Set("X-Real-IP", ip)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	return client.Do(ctx, req)
}
