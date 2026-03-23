package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"reliquary/internal/db"
	"reliquary/internal/handlers"
	"reliquary/internal/middleware"
)
git
func main() {
	if err := db.Init(); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(middleware.Auth)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Route("/api/libraries", func(r chi.Router) {
		r.Get("/", handlers.ListLibraries)
		r.Post("/", handlers.CreateLibrary)
		r.Get("/{id}", handlers.GetLibrary)
		r.Put("/{id}", handlers.UpdateLibrary)
		r.Delete("/{id}", handlers.DeleteLibrary)
	})

	r.Get("/api/directories", handlers.ListDirectories)

	http.ListenAndServe(":5321", r)
}
