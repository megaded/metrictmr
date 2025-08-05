package agent

import (
	"context"

	"github.com/megaded/metrictmr/internal/agent/client"
)

type MetricSender interface {
	StartSend(ctx context.Context)
}

func GetMetricSender() MetricSender {
	return client.CreateProtoClient()
}
