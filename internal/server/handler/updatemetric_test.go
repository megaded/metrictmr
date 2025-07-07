package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
	"github.com/stretchr/testify/assert"
)

func TestSendMetric(t *testing.T) {
	gaugeName := "gauge"
	counterName := "counter"
	tests := []struct {
		name        string
		params      string
		method      string
		contentType string
		code        int
	}{
		{name: "200 store gauge", params: "update/gauge/111/11", code: http.StatusOK, method: http.MethodPost, contentType: "text/plain"},
		{name: "200 store counter", params: "update/counter/111/11", code: http.StatusOK, method: http.MethodPost},
		{name: "200 get gauge", params: fmt.Sprintf("value/gauge/%s", gaugeName), code: http.StatusOK, method: http.MethodGet},
		{name: "200 get counter", params: fmt.Sprintf("value/counter/%s", counterName), code: http.StatusOK, method: http.MethodGet},
		{name: "400 name empty", params: "update/gauge//11", code: http.StatusNotFound, method: http.MethodPost},
		{name: "400 invalid type", params: "update/ffff/11/11", code: http.StatusBadRequest, method: http.MethodPost},
		{name: "400 invalid value", params: "update/gauge/11/fdfdf", code: http.StatusBadRequest, method: http.MethodPost},
	}
	store := storage.NewInMemoryStorage()
	var delta int64 = 1
	var value float64 = 1
	store.Store(context.TODO(), data.Metric{ID: gaugeName, MType: gaugeType, Value: &value}, data.Metric{ID: counterName, MType: counterType, Delta: &delta})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := CreateRouter(store)
			ts := httptest.NewServer(router)
			defer ts.Close()
			client := ts.Client()
			fmt.Println(tt.name + " " + fmt.Sprintf("%v/%v", ts.URL, tt.params))
			request, _ := http.NewRequest(tt.method, fmt.Sprintf("%v/%v", ts.URL, tt.params), nil)
			request.Header.Set("Content-Type", tt.contentType)
			res, err := client.Do(request)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
