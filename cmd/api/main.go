package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"librex/internal/db"
	"librex/internal/handlers"
	"librex/internal/middleware"
)

func main() {
	_ = godotenv.Load(".env.dev")
	_ = godotenv.Load(".env")
	if err := db.Init(); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	// OPDS Routes (Basic Auth)
	r.Route("/opds", func(r chi.Router) {
		r.Use(middleware.OPDSAuth)
		r.Get("/", handlers.GetOPDSRoot)
		r.Get("/all.xml", handlers.GetOPDSAll)
		r.Get("/new.xml", handlers.GetOPDSNew)
		r.Get("/random.xml", handlers.GetOPDSRandom)
		r.Get("/libraries.xml", handlers.GetOPDSLibrariies)
		r.Get("/libraries/{id}.xml", handlers.GetOPDSLibraryBooks)
		r.Get("/shelves.xml", handlers.GetOPDSShelves)
		r.Get("/shelves/{id}.xml", handlers.GetOPDSShelfBooks)
		r.Get("/authors.xml", handlers.GetOPDSAuthors)
		r.Get("/authors/{id}.xml", handlers.GetOPDSAuthorBooks)
		r.Get("/series.xml", handlers.GetOPDSSeries)
		r.Get("/series/{name}.xml", handlers.GetOPDSSeriesBooks)
		r.Get("/magic.xml", handlers.GetOPDSMagic)
		r.Get("/books/{id}/cover", handlers.GetBookCover)
		r.Get("/books/{id}/download", handlers.DownloadBook)
	})

	// Authenticated API Routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)

		r.Route("/api/libraries", func(r chi.Router) {
			r.Get("/", handlers.ListLibraries)
			r.Post("/", handlers.CreateLibrary)
			r.Post("/scan", handlers.ScanAllLibraries)
			r.Get("/{id}", handlers.GetLibrary)
			r.Put("/{id}", handlers.UpdateLibrary)
			r.Delete("/{id}", handlers.DeleteLibrary)
			r.Get("/{id}/books", handlers.ListLibraryBooks)
			r.Post("/{id}/scan", handlers.ScanLibrary)
			r.Post("/{id}/upload", handlers.UploadToLibrary)
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
			r.Delete("/{id}/progress", handlers.DeleteReadingProgress)
			r.Get("/{id}/sessions", handlers.ListReadingSessions)
			r.Post("/{id}/sessions", handlers.CreateReadingSession)
		})

		r.Route("/api/shelves", func(r chi.Router) {
			r.Get("/", handlers.ListShelves)
			r.Post("/", handlers.CreateShelf)
			r.Get("/unshelved/books", handlers.ListUnshelvedBooks)
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

		r.Route("/api/genres", func(r chi.Router) {
			r.Get("/", handlers.ListGenres)
			r.Post("/", handlers.CreateGenre)
			r.Get("/{id}", handlers.GetGenre)
			r.Put("/{id}", handlers.UpdateGenre)
			r.Delete("/{id}", handlers.DeleteGenre)
			r.Get("/{id}/books", handlers.ListGenreBooks)
		})

		r.Route("/api/tags", func(r chi.Router) {
			r.Get("/", handlers.ListTags)
			r.Post("/", handlers.CreateTag)
			r.Get("/{id}", handlers.GetTag)
			r.Put("/{id}", handlers.UpdateTag)
			r.Delete("/{id}", handlers.DeleteTag)
			r.Get("/{id}/books", handlers.ListTagBooks)
		})

		r.Post("/api/books/bulk-update", handlers.BulkUpdateBooks)

		r.Get("/api/series", handlers.ListSeries)
		r.Get("/api/publishers", handlers.ListPublishers)

		r.Route("/api/settings", func(r chi.Router) {
			r.Get("/", handlers.GetSettings)
			r.Put("/", handlers.UpdateSettings)
			r.Get("/opds", handlers.GetOPDSSettings)
			r.Put("/opds", handlers.UpdateOPDSSettings)
		})

		r.Get("/api/directories", handlers.ListDirectories)

		r.Route("/api/bookdrop", func(r chi.Router) {
			r.Post("/scan", handlers.ScanBookdrop)
			r.Post("/upload", handlers.UploadToBookdrop)
			r.Get("/staged", handlers.ListStagedBooks)
			r.Get("/staged/{id}", handlers.GetStagedBook)
			r.Get("/staged/{id}/cover", handlers.GetStagedBookCover)
			r.Put("/staged/{id}", handlers.UpdateStagedBook)
			r.Put("/staged", handlers.BulkUpdateStagedBooks)
			r.Delete("/staged/{id}", handlers.DeleteStagedBook)
			r.Delete("/staged", handlers.ClearStagedBooks)
			r.Post("/import", handlers.ImportBooks)
		})
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "6001"
	}
	log.Printf("Starting server on :%s", port)
	http.ListenAndServe(":"+port, r)
}
