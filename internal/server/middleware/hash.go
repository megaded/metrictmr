package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/megaded/metrictmr/internal/logger"
	"go.uber.org/zap"
)

const HashHeader string = "HashSHA256"

func Hash(key string) func(h http.Handler) http.Handler {
	hFunc := func(h http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {
			hw := w
			if key != "" {
				hashHeader := r.Header.Get(HashHeader)
				if hashHeader == "" {
					hw.WriteHeader(http.StatusBadRequest)
					return
				}
				hashWriter := hashWriter{
					ResponseWriter: w,
					key:            key,
				}
				hw = &hashWriter
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
					logger.Log.Error("Hash not equal", zap.String("request hash", hashHeader), zap.String("calculated hash", hash))
					hw.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			h.ServeHTTP(hw, r)
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
		logger.Log.Info(hash)
		r.ResponseWriter.Header().Set(HashHeader, hash)
	}
	size, err := r.ResponseWriter.Write(b)
	return size, err
}

func (r *hashWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}
