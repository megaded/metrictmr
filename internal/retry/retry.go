package retry

import (
	"context"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/logger"
	"go.uber.org/zap"
)

type Retry struct {
	start    time.Duration
	step     time.Duration
	maxRetry int
}

func NewRetry(start int, step int, maxRetry int) Retry {
	return Retry{start: time.Duration(start * int(time.Second)), step: time.Duration(step * int(time.Second)), maxRetry: maxRetry}
}

func (r *Retry) RetryAgent(ctx context.Context, action func() (*http.Response, error)) func() {
	rt := func() {
		resp, err := action()
		if err != nil {
			logger.Log.Error(err.Error())
			if r.maxRetry <= 0 {
				return
			}
			countRetry := r.maxRetry
			delay := r.start
			t := time.NewTicker(r.start)
			logger.Log.Info("Start retry", zap.Int("retry start", int(r.start)))
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					logger.Log.Info("Retry max", zap.Int("retry max", r.maxRetry))
					logger.Log.Info("Retry count", zap.Int("retry", countRetry))
					delay = delay + r.step
					logger.Log.Info("Next retry", zap.Int("retry next", int(delay)))
					t.Reset(delay)
					countRetry--
					if countRetry == 0 {
						return
					}
				}
			}
		}
		defer resp.Body.Close()
	}
	return rt
}
