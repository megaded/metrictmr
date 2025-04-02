package retry

import (
	"context"
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/logger"
)

type Retry struct {
	start    time.Duration
	step     time.Duration
	maxRetry int
}

func NewRetry(start int, step int, maxRetry int) Retry {
	return Retry{start: time.Duration(start * int(time.Second)), step: time.Duration(step * int(time.Second)), maxRetry: maxRetry}
}

func (r *Retry) RetryAgent(ctx context.Context, action func() (*http.Response, error)) func() error {
	rt := func() error {
		if r.maxRetry <= 0 {
			return nil
		}
		resp, err := action()
		if err != nil {
			logger.Log.Error(err.Error())

			countRetry := 0
			delay := r.start
			t := time.NewTicker(r.start)
			for {
				select {
				case <-ctx.Done():
					return err
				case <-t.C:
					resp, err = action()
					if err != nil {
						delay = delay + r.step
						t.Reset(delay)
						countRetry++
						if countRetry == r.maxRetry {
							return err
						}
					}
					defer resp.Body.Close()
					return nil
				}
			}
		}
		defer resp.Body.Close()
		return err
	}
	return rt
}

func (r *Retry) Retry(ctx context.Context, action func() error) func() error {
	rt := func() error {
		err := action()
		if err != nil {
			if r.maxRetry <= 0 {
				return nil
			}
			logger.Log.Error(err.Error())
			countRetry := 0
			delay := r.start
			t := time.NewTicker(r.start)
			for {
				select {
				case <-ctx.Done():
					return err
				case <-t.C:
					err = action()
					delay = delay + r.step
					t.Reset(delay)
					countRetry++
					if countRetry == r.maxRetry {
						return err
					}
				}
			}
		}
		return err
	}
	return rt
}
