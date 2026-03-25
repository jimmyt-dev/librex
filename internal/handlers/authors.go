package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

type authorWithCount struct {
	models.Author
	BookCount int `json:"bookCount"`
}

// ListAuthors returns all authors for the authenticated user with book counts.
// Supports ?q= for autocomplete filtering.
func ListAuthors(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	q := r.URL.Query().Get("q")
	var query string
	var args []any

	if q != "" {
		query = `SELECT a.id, a.name, a.user_id, COUNT(ba.book_id) AS book_count
			FROM authors a
			LEFT JOIN book_authors ba ON ba.author_id = a.id
			WHERE a.user_id = $1 AND a.name ILIKE $2
			GROUP BY a.id
			ORDER BY a.name`
		args = []any{userID, "%" + q + "%"}
	} else {
		query = `SELECT a.id, a.name, a.user_id, COUNT(ba.book_id) AS book_count
			FROM authors a
			LEFT JOIN book_authors ba ON ba.author_id = a.id
			WHERE a.user_id = $1
			GROUP BY a.id
			ORDER BY a.name`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	authors := []authorWithCount{}
	for rows.Next() {
		var a authorWithCount
		if err := rows.Scan(&a.ID, &a.Name, &a.UserID, &a.BookCount); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		authors = append(authors, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

// GetAuthor returns a single author by ID.
func GetAuthor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var a authorWithCount
	err := db.DB.QueryRow(r.Context(),
		`SELECT a.id, a.name, a.user_id, COUNT(ba.book_id) AS book_count
		FROM authors a
		LEFT JOIN book_authors ba ON ba.author_id = a.id
		WHERE a.id = $1 AND a.user_id = $2
		GROUP BY a.id`, id, userID).
		Scan(&a.ID, &a.Name, &a.UserID, &a.BookCount)
	if err != nil {
		http.Error(w, "author not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

type authorBody struct {
	Name string `json:"name"`
}

// CreateAuthor creates a new author.
func CreateAuthor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body authorBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var a models.Author
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO authors (name, user_id) VALUES ($1, $2)
		ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, user_id`,
		body.Name, userID).Scan(&a.ID, &a.Name, &a.UserID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

// UpdateAuthor renames an author.
func UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var body authorBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var a models.Author
	err := db.DB.QueryRow(r.Context(),
		`UPDATE authors SET name = $1 WHERE id = $2 AND user_id = $3 RETURNING id, name, user_id`,
		body.Name, id, userID).Scan(&a.ID, &a.Name, &a.UserID)
	if err != nil {
		http.Error(w, "author not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

// DeleteAuthor removes an author. Books are not deleted, only the join entries.
func DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM authors WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "author not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListAuthorBooks returns all books by a given author.
func ListAuthorBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	// Verify author belongs to user
	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM authors WHERE id = $1 AND user_id = $2)",
		id, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "author not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_authors ba ON ba.book_id = b.id
		WHERE ba.author_id = $1
		ORDER BY m.title`, id)
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

	if err := attachBookRelations(r, books); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// findOrCreateAuthors takes a list of author names and returns their IDs,
// creating any that don't exist yet.
func findOrCreateAuthors(r *http.Request, names []string, userID string) ([]models.Author, error) {
	authors := make([]models.Author, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var a models.Author
		err := db.DB.QueryRow(r.Context(),
			`INSERT INTO authors (name, user_id) VALUES ($1, $2)
			ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id, name, user_id`,
			name, userID).Scan(&a.ID, &a.Name, &a.UserID)
		if err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, nil
}

// linkBookAuthors replaces all author associations for a book.
func linkBookAuthors(r *http.Request, q db.DBTX, bookID string, authors []models.Author) error {
	_, err := q.Exec(r.Context(), "DELETE FROM book_authors WHERE book_id = $1", bookID)
	if err != nil {
		return err
	}
	for _, a := range authors {
		_, err := q.Exec(r.Context(),
			"INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			bookID, a.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// attachAuthors populates the Authors field on a slice of books.
func attachAuthors(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]string, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}

	rows, err := db.DB.Query(r.Context(),
		`SELECT ba.book_id, a.id, a.name, a.user_id
		FROM book_authors ba
		JOIN authors a ON a.id = ba.author_id
		WHERE ba.book_id = ANY($1)
		ORDER BY a.name`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byBook := map[string][]models.Author{}
	for rows.Next() {
		var bookID string
		var a models.Author
		if err := rows.Scan(&bookID, &a.ID, &a.Name, &a.UserID); err != nil {
			return err
		}
		byBook[bookID] = append(byBook[bookID], a)
	}

	for i := range books {
		if authors, ok := byBook[books[i].ID]; ok {
			books[i].Authors = authors
		} else {
			books[i].Authors = []models.Author{}
		}
	}

	return nil
}

// parseAuthorString splits an author metadata string into individual names.
func parseAuthorString(s string) []string {
	// Split on common delimiters: semicolons, " & ", commas
	// Try semicolons first (most unambiguous)
	var parts []string
	if strings.Contains(s, ";") {
		parts = strings.Split(s, ";")
	} else if strings.Contains(s, " & ") {
		parts = strings.Split(s, " & ")
	} else if strings.Contains(s, ",") {
		parts = strings.Split(s, ",")
	} else {
		parts = []string{s}
	}

	var result []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}
