package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/megaded/metrictmr/internal/server/handler/storage"
)

const (
	gaugeType   = "gauge"
	counterType = "counter"
	typeParam   = "type"
	nameParam   = "name"
)

func CreateRouter(s storage.Storager, middleWare ...func(http.Handler) http.Handler) http.Handler {
	handler := handler{storage: s}
	router := chi.NewRouter()
	router.Use(middleware.Compress(5, "application/json", "text/html"))
	for _, m := range middleWare {
		router.Use(m)
	}
	router.Route("/update", func(r chi.Router) {
		r.Post("/", handler.getSaveJSONHandler())
		r.Post("/{type}/{name}/{value}", handler.getSaveHandler())
	})

	router.Route("/value", func(r chi.Router) {
		r.Post("/", handler.getMetricJSONHandler())
		r.Get("/{type}/{name}", handler.getMetricHandler())
	})

	router.Route("/ping", func(r chi.Router) {
		r.Get("/", handler.getPingDBHandler())
	})

	router.Route("/updates", func(r chi.Router) {
		r.Post("/", handler.getSaveBulkJSONHandler())
	})

	router.Get("/", handler.getMetricListHandler())
	return router
}
