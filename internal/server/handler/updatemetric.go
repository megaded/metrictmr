package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/megaded/metrictmr/internal/server/handler/data"
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
	router.Get("/value/{type}/{name}", getMetricHandler(s))
	router.Get("/", getMetricListHandler(s))
	router.Post("/update/", getSaveJSONHandler(s))
	router.Post("/value/", getMetricJSONHandler(s))
	router.Post("/update/{type}/{name}/{value}", getSaveHandler(s))
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

func getMetricJSONHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric data.Metrics
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		var resp []byte
		switch metric.MType {
		case gaugeType:
			value, ok := s.GetGauge(metric.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Value = &value
			resp, err = json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			value, ok := s.GetCounter(metric.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Delta = &value
			resp, err = json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
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

func getSaveJSONHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric data.Metrics
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		var resp []byte
		switch metric.MType {
		case gaugeType:
			if metric.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.StoreGauge(metric.ID, *metric.Value)
			value, _ := s.GetGauge(metric.ID)
			metric.Value = &value
			resp, err = json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			if metric.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.StoreCounter(metric.ID, *metric.Delta)
			value, _ := s.GetCounter(metric.ID)
			metric.Delta = &value
			resp, err = json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func FloatFormat(value float64) string {
	return strings.TrimRight(fmt.Sprintf("%.3f", value), "0.")
}
