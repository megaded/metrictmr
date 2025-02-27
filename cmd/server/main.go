package main

import (
	"fmt"
	"net/http"

	"github.com/megaded/metrictmr/internal/server/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{type}/{name}/{value}", handler.SendMetric)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
