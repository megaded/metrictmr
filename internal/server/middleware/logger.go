package middleware

import (
	"net/http"
	"time"

	"github.com/megaded/metrictmr/internal/logger"
	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type responseLogWriter struct {
	http.ResponseWriter
	data responseData
}

func (r *responseLogWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func (r *responseLogWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.data.status = statusCode
}

func (r *responseLogWriter) logResponse() {
	logger.Log.Info("Response status", zap.Int("status", r.data.status))
	logger.Log.Info("Response size", zap.Int("size", r.data.status))
}

func Logger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := responseLogWriter{
			ResponseWriter: w,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		lw.logResponse()
		logRequest(r, duration)

	}

	return http.HandlerFunc(logFn)
}

func logRequest(r *http.Request, duration time.Duration) {
	uri := r.RequestURI
	method := r.Method
	logger.Log.Info("Request uri", zap.String("uri", uri))
	logger.Log.Info("Request method", zap.String("method", method))
	logger.Log.Info("Request duration", zap.Duration("duration", duration))
}
