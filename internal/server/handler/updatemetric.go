package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
)

const (
	gaugeType   = "gauge"
	counterType = "counter"
	typeParam   = "type"
	nameParam   = "name"
)

type Storager interface {
	GetGauge(name string) (metric data.Metric, exist bool)
	Store(metric data.Metric)
	GetCounter(name string) (metric data.Metric, exist bool)
	GetGaugeMetrics() []data.Metric
	GetCounterMetrics() []data.Metric
}

func CreateRouter(s Storager, middleWare ...func(http.Handler) http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Compress(5, "application/json", "text/html"))
	for _, m := range middleWare {
		router.Use(m)
	}
	router.Route("/update", func(r chi.Router) {
		r.Post("/", getSaveJSONHandler(s))
		r.Post("/{type}/{name}/{value}", getSaveHandler(s))
	})

	router.Route("/value", func(r chi.Router) {
		r.Post("/", getMetricJSONHandler(s))
		r.Get("/{type}/{name}", getMetricHandler(s))
	})

	router.Get("/", getMetricListHandler(s))
	return router
}

func getMetricListHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := new(bytes.Buffer)
		gaugeMetrics := s.GetGaugeMetrics()
		for _, value := range gaugeMetrics {
			fmt.Fprintf(b, "Name %v=\"%f\"\n", value.ID, *value.Value)
		}
		counterMetrics := s.GetCounterMetrics()
		for _, value := range counterMetrics {
			fmt.Fprintf(b, "Name %v=\"%d\"\n", value.ID, *value.Delta)
		}
		w.Header().Set("Content-Type", "text/html")
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
				w.Write([]byte(FloatFormat(*value.Value)))
				return
			}
		case counterType:
			value, ok := s.GetCounter(mName)
			if ok {
				w.Write([]byte(fmt.Sprintf("%d", *value.Delta)))
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
	}
}

func getMetricJSONHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric data.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		var resp []byte
		switch metric.MType {
		case gaugeType:
			storedMetric, ok := s.GetGauge(metric.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Value = storedMetric.Value
			resp, err = json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			storedMetric, ok := s.GetCounter(metric.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metric.Delta = storedMetric.Delta
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
			s.Store(data.Metric{ID: mName, MType: gaugeType, Value: &fValue})
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
			s.Store(data.Metric{ID: mName, MType: counterType, Delta: &fValue})
		}

		w.WriteHeader(statusCode)
	}
}

func getSaveJSONHandler(s Storager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric data.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			logger.Log.Info(string(bodyBytes))
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Info(err.Error())
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
			s.Store(metric)
			storedMetric, _ := s.GetGauge(metric.ID)
			resp, err = json.Marshal(storedMetric)
			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			if metric.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.Store(metric)
			storedMetric, _ := s.GetCounter(metric.ID)
			resp, err = json.Marshal(storedMetric)
			if err != nil {
				logger.Log.Info(err.Error())
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
