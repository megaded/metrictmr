// Методы для работы с метриками
// Поддерживает два типа метрик gauge и counter
//
// gauge чисто с плавающей точной
// counter целое положительное число
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
)

type handler struct {
	storage storage.Storager
}

// Возвращает страницу со списком метрик
func (h *handler) getMetricListHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := new(bytes.Buffer)
		metrics, err := h.storage.GetMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, value := range metrics {
			if value.MType == gaugeType {
				fmt.Fprintf(b, "Name %v=\"%f\"\n", value.ID, *value.Value)
			}
			if value.MType == counterType {
				fmt.Fprintf(b, "Name %v=\"%d\"\n", value.ID, *value.Delta)
			}

		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	}
}

// Пинг к БД
func (h *handler) getPingDBHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ping := h.storage.HealthCheck()
		if ping {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

}

// Получение метрики по имени
func (h *handler) getMetricHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, typeParam)
		if mType != gaugeType && mType != counterType {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mName := chi.URLParam(r, nameParam)
		switch mType {
		case gaugeType:
			value, ok, err := h.storage.GetGauge(mName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if ok {
				w.Write([]byte(FloatFormat(*value.Value)))
				return
			}
		case counterType:
			value, ok, err := h.storage.GetCounter(mName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if ok {
				w.Write([]byte(fmt.Sprintf("%d", *value.Delta)))
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
	}
}

// Получение метрики по имени
func (h *handler) getMetricJSONHandler() func(w http.ResponseWriter, r *http.Request) {
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
			storedMetric, ok, err := h.storage.GetGauge(metric.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
			storedMetric, ok, err := h.storage.GetCounter(metric.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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

// Сохранение метрики
func (h *handler) getSaveHandler() func(w http.ResponseWriter, r *http.Request) {
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
			fValue, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				statusCode = http.StatusBadRequest
				break
			}
			if fValue < 0 {
				statusCode = http.StatusBadRequest
				break
			}
			err = h.storage.Store(r.Context(), data.Metric{ID: mName, MType: gaugeType, Value: &fValue})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			fValue, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				statusCode = http.StatusBadRequest
				break
			}
			if fValue < 0 {
				statusCode = http.StatusBadRequest
				break
			}
			err = h.storage.Store(r.Context(), data.Metric{ID: mName, MType: counterType, Delta: &fValue})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(statusCode)
	}
}

// Сохранение метрики
func (h *handler) getSaveJSONHandler() func(w http.ResponseWriter, r *http.Request) {
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
			if metric.Value == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = h.storage.Store(r.Context(), metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			storedMetric, _, err := h.storage.GetGauge(metric.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			resp, err = json.Marshal(storedMetric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case counterType:
			if metric.Delta == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = h.storage.Store(r.Context(), metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			storedMetric, _, err := h.storage.GetCounter(metric.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			resp, err = json.Marshal(storedMetric)
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

// Сохранение списка метрик
func (h *handler) getSaveBulkJSONHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric []data.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = h.storage.Store(r.Context(), metric...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func FloatFormat(value float64) string {
	return strings.TrimRight(fmt.Sprintf("%.3f", value), "0.")
}
