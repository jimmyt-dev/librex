package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

const bookCols = `id, library_id, user_id, title, author, subject, description, publisher, contributor,
	date, type, format, identifier, source, language, relation, coverage,
	cover_image IS NOT NULL AS has_cover,
	file_path`

func scanBook(scan func(dest ...any) error) (models.Book, error) {
	var b models.Book
	var hasCover bool
	err := scan(
		&b.ID, &b.LibraryID, &b.UserID, &b.Title, &b.Author,
		&b.Subject, &b.Description, &b.Publisher, &b.Contributor,
		&b.Date, &b.Type, &b.Format, &b.Identifier,
		&b.Source, &b.Language, &b.Relation, &b.Coverage,
		&hasCover, &b.FilePath,
	)
	if err == nil && hasCover {
		url := fmt.Sprintf("/api/books/%s/cover", b.ID)
		b.Cover = &url
	}
	return b, err
}

// ListLibraryBooks returns all books in a library.
func ListLibraryBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	libraryID := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM libraries WHERE id = $1 AND user_id = $2)",
		libraryID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		"SELECT "+bookCols+" FROM books WHERE library_id = $1 ORDER BY title", libraryID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetBook returns a single book by ID.
func GetBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		"SELECT "+bookCols+" FROM books WHERE id = $1 AND user_id = $2", id, userID)
	b, err := scanBook(row.Scan)
	if err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func GetBookAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := db.DB.Query(r.Context(),
		"SELECT "+bookCols+" FROM books WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "books not found", http.StatusNotFound)
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetBookCover serves the cover image for a book.
func GetBookCover(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var imgData []byte
	var mime string
	err := db.DB.QueryRow(r.Context(),
		"SELECT cover_image, cover_mime FROM books WHERE id = $1 AND user_id = $2",
		id, userID).Scan(&imgData, &mime)
	if err != nil || len(imgData) == 0 {
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	}

	if mime == "" {
		mime = "image/jpeg"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(imgData)
}

type bookUpdate struct {
	Title       *string `json:"title"`
	Author      *string `json:"author"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	Publisher   *string `json:"publisher"`
	Contributor *string `json:"contributor"`
	Date        *string `json:"date"`
	Type        *string `json:"type"`
	Format      *string `json:"format"`
	Identifier  *string `json:"identifier"`
	Source      *string `json:"source"`
	Language    *string `json:"language"`
	Relation    *string `json:"relation"`
	Coverage    *string `json:"coverage"`
}

// UpdateBook updates the metadata of a book.
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		"SELECT "+bookCols+" FROM books WHERE id = $1 AND user_id = $2", id, userID)
	existing, err := scanBook(row.Scan)
	if err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	var body bookUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title != nil {
		existing.Title = *body.Title
	}
	if body.Author != nil {
		existing.Author = body.Author
	}
	if body.Subject != nil {
		existing.Subject = body.Subject
	}
	if body.Description != nil {
		existing.Description = body.Description
	}
	if body.Publisher != nil {
		existing.Publisher = body.Publisher
	}
	if body.Contributor != nil {
		existing.Contributor = body.Contributor
	}
	if body.Date != nil {
		existing.Date = body.Date
	}
	if body.Type != nil {
		existing.Type = body.Type
	}
	if body.Format != nil {
		existing.Format = body.Format
	}
	if body.Identifier != nil {
		existing.Identifier = body.Identifier
	}
	if body.Source != nil {
		existing.Source = body.Source
	}
	if body.Language != nil {
		existing.Language = body.Language
	}
	if body.Relation != nil {
		existing.Relation = body.Relation
	}
	if body.Coverage != nil {
		existing.Coverage = body.Coverage
	}

	_, err = db.DB.Exec(r.Context(),
		`UPDATE books SET
			title=$1, author=$2, subject=$3, description=$4, publisher=$5, contributor=$6,
			date=$7, type=$8, format=$9, identifier=$10, source=$11, language=$12,
			relation=$13, coverage=$14
		WHERE id = $15 AND user_id = $16`,
		existing.Title, existing.Author,
		existing.Subject, existing.Description, existing.Publisher, existing.Contributor,
		existing.Date, existing.Type, existing.Format, existing.Identifier,
		existing.Source, existing.Language, existing.Relation, existing.Coverage,
		id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// DeleteBook removes a book record. Pass ?deleteFile=true to also delete the file from disk.
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	shouldDeleteFile := r.URL.Query().Get("deleteFile") == "true"

	var filePath string
	if shouldDeleteFile {
		if err := db.DB.QueryRow(r.Context(),
			"SELECT file_path FROM books WHERE id = $1 AND user_id = $2", id, userID).Scan(&filePath); err != nil {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
	}

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM books WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	if shouldDeleteFile && filePath != "" {
		os.Remove(filePath)
	}

	w.WriteHeader(http.StatusNoContent)
}
