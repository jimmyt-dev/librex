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
		r.Post("/scan", handlers.ScanAllLibraries)
		r.Get("/{id}", handlers.GetLibrary)
		r.Put("/{id}", handlers.UpdateLibrary)
		r.Delete("/{id}", handlers.DeleteLibrary)
		r.Get("/{id}/books", handlers.ListLibraryBooks)
		r.Post("/{id}/scan", handlers.ScanLibrary)
	})

	r.Route("/api/books", func(r chi.Router) {
		r.Get("/all", handlers.GetBookAll)
		r.Post("/move", handlers.MoveBooks)
		r.Get("/{id}", handlers.GetBook)
		r.Put("/{id}", handlers.UpdateBook)
		r.Delete("/{id}", handlers.DeleteBook)
		r.Get("/{id}/cover", handlers.GetBookCover)
		r.Get("/{id}/download", handlers.DownloadBook)
		r.Get("/{id}/shelves", handlers.ListBookShelves)
		r.Get("/{id}/progress", handlers.GetReadingProgress)
		r.Put("/{id}/progress", handlers.UpdateReadingProgress)
		r.Get("/{id}/sessions", handlers.ListReadingSessions)
		r.Post("/{id}/sessions", handlers.CreateReadingSession)
	})

	r.Route("/api/shelves", func(r chi.Router) {
		r.Get("/", handlers.ListShelves)
		r.Post("/", handlers.CreateShelf)
		r.Get("/unshelved", handlers.ListUnshelvedBooks)
		r.Get("/{id}", handlers.GetShelf)
		r.Put("/{id}", handlers.UpdateShelf)
		r.Delete("/{id}", handlers.DeleteShelf)
		r.Get("/{id}/books", handlers.ListShelfBooks)
		r.Post("/{id}/books", handlers.AddBooksToShelf)
		r.Delete("/{id}/books", handlers.RemoveBooksFromShelf)
	})

	r.Route("/api/authors", func(r chi.Router) {
		r.Get("/", handlers.ListAuthors)
		r.Post("/", handlers.CreateAuthor)
		r.Get("/{id}", handlers.GetAuthor)
		r.Put("/{id}", handlers.UpdateAuthor)
		r.Delete("/{id}", handlers.DeleteAuthor)
		r.Get("/{id}/books", handlers.ListAuthorBooks)
	})

	r.Route("/api/categories", func(r chi.Router) {
		r.Get("/", handlers.ListCategories)
		r.Post("/", handlers.CreateCategory)
		r.Get("/{id}", handlers.GetCategory)
		r.Put("/{id}", handlers.UpdateCategory)
		r.Delete("/{id}", handlers.DeleteCategory)
		r.Get("/{id}/books", handlers.ListCategoryBooks)
	})

	r.Route("/api/tags", func(r chi.Router) {
		r.Get("/", handlers.ListTags)
		r.Post("/", handlers.CreateTag)
		r.Get("/{id}", handlers.GetTag)
		r.Put("/{id}", handlers.UpdateTag)
		r.Delete("/{id}", handlers.DeleteTag)
		r.Get("/{id}/books", handlers.ListTagBooks)
	})

	r.Get("/api/series", handlers.ListSeries)

	r.Route("/api/settings", func(r chi.Router) {
		r.Get("/", handlers.GetSettings)
		r.Put("/", handlers.UpdateSettings)
	})

	r.Get("/api/directories", handlers.ListDirectories)

	r.Route("/api/bookdrop", func(r chi.Router) {
		r.Post("/scan", handlers.ScanBookdrop)
		r.Get("/staged", handlers.ListStagedBooks)
		r.Get("/staged/{id}", handlers.GetStagedBook)
		r.Get("/staged/{id}/cover", handlers.GetStagedBookCover)
		r.Put("/staged/{id}", handlers.UpdateStagedBook)
		r.Put("/staged", handlers.BulkUpdateStagedBooks)
		r.Delete("/staged/{id}", handlers.DeleteStagedBook)
		r.Delete("/staged", handlers.ClearStagedBooks)
		r.Post("/import", handlers.ImportBooks)
	})

	http.ListenAndServe(":5321", r)
}
