package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

type categoryWithCount struct {
	models.Category
	BookCount int `json:"bookCount"`
}

// ListCategories returns all categories for the authenticated user with book counts.
func ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	q := r.URL.Query().Get("q")
	var rows_query string
	var args []any

	if q != "" {
		rows_query = `SELECT c.id, c.name, c.user_id, COUNT(bc.book_id) AS book_count
			FROM categories c
			LEFT JOIN book_categories bc ON bc.category_id = c.id
			WHERE c.user_id = $1 AND c.name ILIKE $2
			GROUP BY c.id
			ORDER BY c.name`
		args = []any{userID, "%" + q + "%"}
	} else {
		rows_query = `SELECT c.id, c.name, c.user_id, COUNT(bc.book_id) AS book_count
			FROM categories c
			LEFT JOIN book_categories bc ON bc.category_id = c.id
			WHERE c.user_id = $1
			GROUP BY c.id
			ORDER BY c.name`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), rows_query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []categoryWithCount{}
	for rows.Next() {
		var c categoryWithCount
		if err := rows.Scan(&c.ID, &c.Name, &c.UserID, &c.BookCount); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetCategory returns a single category by ID.
func GetCategory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var c categoryWithCount
	err := db.DB.QueryRow(r.Context(),
		`SELECT c.id, c.name, c.user_id, COUNT(bc.book_id) AS book_count
		FROM categories c
		LEFT JOIN book_categories bc ON bc.category_id = c.id
		WHERE c.id = $1 AND c.user_id = $2
		GROUP BY c.id`, id, userID).
		Scan(&c.ID, &c.Name, &c.UserID, &c.BookCount)
	if err != nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

type categoryBody struct {
	Name string `json:"name"`
}

// CreateCategory creates a new category.
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body categoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var c models.Category
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO categories (name, user_id) VALUES ($1, $2)
		ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, user_id`,
		body.Name, userID).Scan(&c.ID, &c.Name, &c.UserID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// UpdateCategory renames a category.
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var body categoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var c models.Category
	err := db.DB.QueryRow(r.Context(),
		`UPDATE categories SET name = $1 WHERE id = $2 AND user_id = $3 RETURNING id, name, user_id`,
		body.Name, id, userID).Scan(&c.ID, &c.Name, &c.UserID)
	if err != nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// DeleteCategory removes a category.
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM categories WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListCategoryBooks returns all books in a given category.
func ListCategoryBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)",
		id, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_categories bc ON bc.book_id = b.id
		WHERE bc.category_id = $1
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
