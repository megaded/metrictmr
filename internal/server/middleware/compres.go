package middleware

import "net/http"

type Compressor interface {
	SetCompressor(w *http.ResponseWriter, r *http.Request)
	Compress()
}

func GetCompressMiddleware(c Compressor) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			c.SetCompressor(&w, r)
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(logFn)
	}
}
