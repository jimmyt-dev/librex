package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"reliquary/internal/db"
	"reliquary/internal/handlers"
	"reliquary/internal/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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

	r.Route("/api/shelves", func(r chi.Router) {
		r.Get("/", handlers.ListShelves)
		r.Post("/", handlers.CreateShelf)
		r.Get("/{id}", handlers.GetShelf)
		r.Put("/{id}", handlers.UpdateShelf)
		r.Delete("/{id}", handlers.DeleteShelf)
	})

	r.Get("/api/directories", handlers.ListDirectories)

	r.Route("/api/bookdrop", func(r chi.Router) {
		r.Post("/scan", handlers.ScanBookdrop)
		r.Get("/staged", handlers.ListStagedBooks)
		r.Get("/staged/{id}", handlers.GetStagedBook)
		r.Put("/staged/{id}", handlers.UpdateStagedBook)
		r.Put("/staged", handlers.BulkUpdateStagedBooks)
		r.Delete("/staged/{id}", handlers.DeleteStagedBook)
		r.Delete("/staged", handlers.ClearStagedBooks)
	})

	http.ListenAndServe(":5321", r)
}
