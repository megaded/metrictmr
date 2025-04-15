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

func HashWriter(key string) func(h http.Handler) http.Handler {
	hFunc := func(h http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {
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
			h.ServeHTTP(w, r)

		}
		return http.HandlerFunc(hashFn)
	}
	return hFunc
}

func HashReader(key string) func(h http.Handler) http.Handler {
	hFunc := func(h http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {

			lw := hashReader{
				ResponseWriter: w,
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
			h.ServeHTTP(&lw, r)
		}
		return http.HandlerFunc(hashFn)
	}
	return hFunc
}

type hashReader struct {
	http.ResponseWriter
	key string
}

func (r *hashReader) Write(b []byte) (int, error) {
	if r.key != "" {
		h := hmac.New(sha256.New, []byte(r.key))
		h.Write(b)
		hash := hex.EncodeToString(h.Sum(nil))
		r.ResponseWriter.Header().Add(hashHeader, hash)
	}
	size, err := r.ResponseWriter.Write(b)
	return size, err
}

func (r *hashReader) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}
