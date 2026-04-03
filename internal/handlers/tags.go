package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"librex/internal/db"
	"librex/internal/middleware"
	"librex/internal/models"
)

type tagWithCount struct {
	models.Tag
	BookCount int `json:"bookCount"`
}

// ListTags returns all tags for the authenticated user with book counts.
func ListTags(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	q := r.URL.Query().Get("q")
	var rows_query string
	var args []any

	if q != "" {
		rows_query = `SELECT t.id, t.name, t.user_id, COUNT(bt.book_id) AS book_count
			FROM tags t
			LEFT JOIN book_tags bt ON bt.tag_id = t.id
			WHERE t.user_id = $1 AND t.name ILIKE $2
			GROUP BY t.id
			ORDER BY t.name`
		args = []any{userID, "%" + q + "%"}
	} else {
		rows_query = `SELECT t.id, t.name, t.user_id, COUNT(bt.book_id) AS book_count
			FROM tags t
			LEFT JOIN book_tags bt ON bt.tag_id = t.id
			WHERE t.user_id = $1
			GROUP BY t.id
			ORDER BY t.name`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), rows_query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tags := []tagWithCount{}
	for rows.Next() {
		var t tagWithCount
		if err := rows.Scan(&t.ID, &t.Name, &t.UserID, &t.BookCount); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		tags = append(tags, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// GetTag returns a single tag by ID.
func GetTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var t tagWithCount
	err := db.DB.QueryRow(r.Context(),
		`SELECT t.id, t.name, t.user_id, COUNT(bt.book_id) AS book_count
		FROM tags t
		LEFT JOIN book_tags bt ON bt.tag_id = t.id
		WHERE t.id = $1 AND t.user_id = $2
		GROUP BY t.id`, id, userID).
		Scan(&t.ID, &t.Name, &t.UserID, &t.BookCount)
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

type tagBody struct {
	Name string `json:"name"`
}

// CreateTag creates a new tag.
func CreateTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body tagBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var t models.Tag
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO tags (name, user_id) VALUES ($1, $2)
		ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, user_id`,
		body.Name, userID).Scan(&t.ID, &t.Name, &t.UserID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// UpdateTag renames a tag.
func UpdateTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var body tagBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var t models.Tag
	err := db.DB.QueryRow(r.Context(),
		`UPDATE tags SET name = $1 WHERE id = $2 AND user_id = $3 RETURNING id, name, user_id`,
		body.Name, id, userID).Scan(&t.ID, &t.Name, &t.UserID)
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// DeleteTag removes a tag.
func DeleteTag(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM tags WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTagBooks returns all books with a given tag.
func ListTagBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1 AND user_id = $2)",
		id, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_tags bt ON bt.book_id = b.id
		WHERE bt.tag_id = $1
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
