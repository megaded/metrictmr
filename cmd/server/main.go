package main

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	gaugeType   = "gauge"
	counterType = "counter"
)

type MemStorage struct {
}

func sendMetric(w http.ResponseWriter, r *http.Request) {
	mName := r.PathValue("name")
	fmt.Println("Имя" + mName)
	if mName == "" {
		fmt.Println("Name r")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	mType := r.PathValue("type")
	if mType != gaugeType && mType != counterType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mValue := r.PathValue("value")
	if mValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
	}
	w.WriteHeader(statusCode)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{type}/{name}/{value}", sendMetric)
	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
