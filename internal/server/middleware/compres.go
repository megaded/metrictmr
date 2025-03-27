package middleware

import (
	"net/http"
	"strings"
)

type Compressor interface {
	GetCompressor(w http.ResponseWriter, r *http.Request) http.ResponseWriter
	Compress()
}

func GetCompressMiddleware(c Compressor) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			gzip := c.GetCompressor(w, r)
			h.ServeHTTP(gzip, r)
		}
		return http.HandlerFunc(logFn)
	}
}

func GzipMiddleware(h http.Handler) http.Handler {
	comFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
		}

		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(comFn)
}
