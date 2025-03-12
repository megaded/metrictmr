package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	gaugeType   = "gauge"
	counterType = "counter"
	typeParam   = "type"
	nameParam   = "name"
)

type Storager interface {
	GetGauge(name string) (value float64, exist bool)
	StoreGauge(name string, value float64)
	GetCounter(name string) (value int64, exist bool)
	StoreCounter(name string, value int64)
	GetGaugeMetrics() map[string]float64
	GetCounterMetrics() map[string]int64
}

func CreateRouter(s Storager, middleWare ...func(http.Handler) http.Handler) http.Handler {
	router := chi.NewRouter()
	for _, m := range middleWare {
		router.Use(m)
	}
	storeHandler := getSaveHandler(s)
	getHandler := getMetricHandler(s)
	getListHandler := getMetricListHandler(s)
	router.Post("/update/{type}/{name}/{value}", storeHandler)
	router.Get("/value/{type}/{name}", getHandler)
	router.Get("/", getListHandler)

	return router
}

func getMetricListHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := new(bytes.Buffer)
		gaugeMetrics := s.GetGaugeMetrics()
		for key, value := range gaugeMetrics {
			fmt.Fprintf(b, "Name %v=\"%f\"\n", key, value)
		}
		counterMetrics := s.GetCounterMetrics()
		for key, value := range counterMetrics {
			fmt.Fprintf(b, "Name %v=\"%d\"\n", key, value)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	}
}

func getMetricHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, typeParam)
		if mType != gaugeType && mType != counterType {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mName := chi.URLParam(r, nameParam)
		switch mType {
		case gaugeType:
			value, ok := s.GetGauge(mName)
			if ok {
				w.Write([]byte(FloatFormat(value)))
				return
			}
		case counterType:
			value, ok := s.GetCounter(mName)
			if ok {
				w.Write([]byte(fmt.Sprintf("%d", value)))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func getSaveHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, nameParam)
		if mName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mType := chi.URLParam(r, typeParam)
		if mType != gaugeType && mType != counterType {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mValue := chi.URLParam(r, "value")
		statusCode := http.StatusOK
		switch mType {
		case gaugeType:
			fValue, error := strconv.ParseFloat(mValue, 64)
			if error != nil {
				statusCode = http.StatusBadRequest
				break
			}
			if fValue <= 0 {
				statusCode = http.StatusBadRequest
				break
			}
			s.StoreGauge(mName, fValue)
		case counterType:
			fValue, error := strconv.ParseInt(mValue, 10, 64)
			if error != nil {
				statusCode = http.StatusBadRequest
				break
			}
			if fValue <= 0 {
				statusCode = http.StatusBadRequest
				break
			}
			s.StoreCounter(mName, fValue)
		}

		w.WriteHeader(statusCode)
	}
}

func FloatFormat(value float64) string {
	return strings.TrimRight(fmt.Sprintf("%.3f", value), "0.")
}
