package main

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	r := chi.NewRouter()

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "Hello, world!")
	})

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metric-type}", func(r chi.Router) {
			r.Route("/{metric-name}", func(r chi.Router) {
				r.Post("/{metric-value}", func(rw http.ResponseWriter, r *http.Request) {
					// тут работа с метрикой
				})
			})
		})
	})

	// r передаётся как http.Handler
	http.ListenAndServe(":8080", r)

}
