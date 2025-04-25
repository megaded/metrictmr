package retry

import (
	"context"
	"errors"
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

func (r *Retry) RetryAgent(ctx context.Context, action func() (*http.Response, error)) func() error {
	rt := func() error {
		delay := r.start
		for attempt := 0; attempt <= r.maxRetry; attempt++ {
			resp, err := action()
			if err == nil {
				defer resp.Body.Close()
				return nil
			}

			if attempt == r.maxRetry {
				return err
			}

			logger.Log.Error("retry failed", zap.Int("attempt", attempt), zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.step):
				delay += r.step
			}
		}
		return errors.New("unreachable retry state")
	}
	return rt
}

func NewRetry(start int, step int, maxRetry int) Retry {
	return Retry{start: time.Duration(start * int(time.Second)), step: time.Duration(step * int(time.Second)), maxRetry: maxRetry}
}

func (r *Retry) Retry(ctx context.Context, action func() error) func() error {
	rt := func() error {
		delay := r.start
		for attempt := 0; attempt <= r.maxRetry; attempt++ {
			err := action()
			if err == nil {
				return nil
			}

			if attempt == r.maxRetry {
				return err
			}

			logger.Log.Error("retry failed", zap.Int("attempt", attempt), zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.step):
				delay += r.step
			}
		}
		return errors.New("unreachable retry state")
	}
	return rt
}
