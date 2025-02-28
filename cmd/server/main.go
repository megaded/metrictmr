package main

import (
	"fmt"
	"net/http"

	"github.com/megaded/metrictmr/internal/server/handler"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
)

func main() {
	s := storage.NewStorage()
	router := handler.CreateRouter(s)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
