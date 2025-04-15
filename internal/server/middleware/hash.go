package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/megaded/metrictmr/internal/logger"
)

const hashHeader string = "HashSHA256"

func Hash(key string) func(h http.Handler) http.Handler {
	hFunc := func(h http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {
			hashWriter := hashWriter{
				ResponseWriter: w,
				key:            key,
			}
			hashHeader := r.Header.Get(hashHeader)
			if r.Body != nil && hashHeader != "" && key != "" {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Log.Info(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				h := hmac.New(sha256.New, []byte(key))
				h.Write(bodyBytes)
				hash := hex.EncodeToString(h.Sum(nil))
				if hash != hashHeader {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			h.ServeHTTP(&hashWriter, r)
		}
		return http.HandlerFunc(hashFn)
	}
	return hFunc
}

type hashWriter struct {
	http.ResponseWriter
	key string
}

func (r *hashWriter) Write(b []byte) (int, error) {
	if r.key != "" {
		h := hmac.New(sha256.New, []byte(r.key))
		h.Write(b)
		hash := hex.EncodeToString(h.Sum(nil))
		r.ResponseWriter.Header().Add(hashHeader, hash)
	}
	size, err := r.ResponseWriter.Write(b)
	return size, err
}

func (r *hashWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}
