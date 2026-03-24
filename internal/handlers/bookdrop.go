package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/metadata"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

var validBookExts = map[string]bool{
	".epub": true,
	".pdf":  true,
	".mobi": true,
	".azw3": true,
	".cbz":  true,
	".cbr":  true,
}

// ScanBookdrop scans the bookdrop directory and inserts new files into staged_books.
func ScanBookdrop(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	targetDir := r.URL.Query().Get("path")
	if targetDir == "" {
		targetDir = "/Users/jimmy/Documents/Code Shit/reliquary/data/bookdrop"
	}

	cleanedDir := filepath.Clean(targetDir)

	if !filepath.IsAbs(cleanedDir) {
		http.Error(w, "bookdrop path must be absolute", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(cleanedDir)
	if err != nil || !info.IsDir() {
		http.Error(w, "bookdrop directory does not exist or is not a directory", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(cleanedDir)
	if err != nil {
		http.Error(w, "failed to read bookdrop directory", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !validBookExts[ext] {
			continue
		}

		baseName := entry.Name()
		originalPath := filepath.Join(cleanedDir, baseName)

		var count int
		if err := db.DB.QueryRow(r.Context(),
			"SELECT COUNT(*) FROM staged_books WHERE original_path = $1 AND user_id = $2",
			originalPath, userID).Scan(&count); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if count > 0 {
			continue
		}

		// Extract metadata from the file (title, author)
		meta := metadata.Extract(originalPath)
		title := meta.Title
		if title == "" {
			title = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		}
		var author *string
		if meta.Author != "" {
			author = &meta.Author
		}

		_, err = db.DB.Exec(r.Context(),
			"INSERT INTO staged_books (title, author, file_name, ext, original_path, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
			title, author, baseName, ext, originalPath, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	listStagedBooks(w, r, userID)
}

// ListStagedBooks returns all staged books for the authenticated user.
func ListStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	listStagedBooks(w, r, userID)
}

func listStagedBooks(w http.ResponseWriter, r *http.Request, userID string) {
	rows, err := db.DB.Query(r.Context(),
		"SELECT id, title, author, file_name, ext, original_path, user_id FROM staged_books WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []models.StagedBook{}
	for rows.Next() {
		var b models.StagedBook
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.FileName, &b.Ext, &b.OriginalPath, &b.UserID); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetStagedBook returns a single staged book by ID.
func GetStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var b models.StagedBook
	err := db.DB.QueryRow(r.Context(),
		"SELECT id, title, author, file_name, ext, original_path, user_id FROM staged_books WHERE id = $1 AND user_id = $2",
		id, userID).Scan(&b.ID, &b.Title, &b.Author, &b.FileName, &b.Ext, &b.OriginalPath, &b.UserID)
	if err != nil {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

type stagedBookUpdate struct {
	Title  *string `json:"title"`
	Author *string `json:"author"`
}

// UpdateStagedBook updates the editable metadata (title, author) of a staged book.
func UpdateStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var existing models.StagedBook
	err := db.DB.QueryRow(r.Context(),
		"SELECT id, title, author, file_name, ext, original_path, user_id FROM staged_books WHERE id = $1 AND user_id = $2",
		id, userID).Scan(&existing.ID, &existing.Title, &existing.Author, &existing.FileName, &existing.Ext, &existing.OriginalPath, &existing.UserID)
	if err != nil {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	var body stagedBookUpdate
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

	_, err = db.DB.Exec(r.Context(),
		"UPDATE staged_books SET title = $1, author = $2 WHERE id = $3 AND user_id = $4",
		existing.Title, existing.Author, id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

type bulkUpdateItem struct {
	ID     string  `json:"id"`
	Title  *string `json:"title"`
	Author *string `json:"author"`
}

// BulkUpdateStagedBooks updates multiple staged books at once.
func BulkUpdateStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var items []bulkUpdateItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	for _, item := range items {
		if item.Title != nil && item.Author != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET title = $1, author = $2 WHERE id = $3 AND user_id = $4",
				*item.Title, *item.Author, item.ID, userID)
		} else if item.Title != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET title = $1 WHERE id = $2 AND user_id = $3",
				*item.Title, item.ID, userID)
		} else if item.Author != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET author = $1 WHERE id = $2 AND user_id = $3",
				*item.Author, item.ID, userID)
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	listStagedBooks(w, r, userID)
}

// DeleteStagedBook removes a staged book by ID.
func DeleteStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM staged_books WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearStagedBooks removes all staged books for the authenticated user.
func ClearStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	_, err := db.DB.Exec(r.Context(),
		"DELETE FROM staged_books WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
