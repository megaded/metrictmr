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
	return &z
}

func (c *GZipCompressor) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *GZipCompressor) Write(p []byte) (int, error) {
	return c.ResponseWriter.Write(p)
}

func (c *GZipCompressor) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
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
