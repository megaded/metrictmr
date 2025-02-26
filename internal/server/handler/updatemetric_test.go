package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name   string
		params string
		code   int
	}{
		{name: "200 gauge", params: "gauge/111/111", code: http.StatusOK},
		{name: "200 counter", params: "counter/111/11", code: http.StatusOK},
		{name: "400 name empty", params: "gauge//11", code: http.StatusNotFound},
		{name: "400 invalid type", params: "ffff/11/11", code: http.StatusBadRequest},
		{name: "400 invalid value", params: "gauge/11/fdfdf", code: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := http.NewServeMux()
			router.HandleFunc("POST /update/{type}/{name}/{value}", SendMetric)
			ts := httptest.NewServer(router)
			defer ts.Close()
			client := ts.Client()
			fmt.Println(tt.name + fmt.Sprintf("%v/update/%v", ts.URL, tt.params))
			res, err := client.Post(fmt.Sprintf("%v/update/%v", ts.URL, tt.params), "text/plain", nil)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
