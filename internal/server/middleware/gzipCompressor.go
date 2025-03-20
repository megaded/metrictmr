package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type GZipCompressor struct {
	http.ResponseWriter
	http.Request
}

func (z GZipCompressor) GetCompressor(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	ow := w
	acceptEncoding := r.Header.Get("Accept-Encoding")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")
	if supportsGzip {
		cw := newCompressWriter(w)
		ow = cw
		defer cw.Close()
	}
	z.ResponseWriter = ow
	z.Request = *r
	return z.ResponseWriter
}

func (z GZipCompressor) Compress() {
	contentEncoding := z.Request.Header.Get("Content-Encoding")
	sendsGzip := strings.Contains(contentEncoding, "gzip")
	if sendsGzip {
		cr, err := newCompressReader(z.Request.Body)
		if err != nil {
			z.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		z.Request.Body = cr
		defer cr.Close()
	}
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
