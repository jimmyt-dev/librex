package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

type genreWithCount struct {
	models.Genre
	BookCount int `json:"bookCount"`
}

// ListGenres returns all genres for the authenticated user with book counts.
func ListGenres(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	q := r.URL.Query().Get("q")
	var rowsQuery string
	var args []any

	if q != "" {
		rowsQuery = `SELECT g.id, g.name, g.user_id, COUNT(bg.book_id) AS book_count
			FROM genres g
			LEFT JOIN book_genres bg ON bg.genre_id = g.id
			WHERE g.user_id = $1 AND g.name ILIKE $2
			GROUP BY g.id
			ORDER BY g.name`
		args = []any{userID, "%" + q + "%"}
	} else {
		rowsQuery = `SELECT g.id, g.name, g.user_id, COUNT(bg.book_id) AS book_count
			FROM genres g
			LEFT JOIN book_genres bg ON bg.genre_id = g.id
			WHERE g.user_id = $1
			GROUP BY g.id
			ORDER BY g.name`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), rowsQuery, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	genres := []genreWithCount{}
	for rows.Next() {
		var g genreWithCount
		if err := rows.Scan(&g.ID, &g.Name, &g.UserID, &g.BookCount); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		genres = append(genres, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

// GetGenre returns a single genre by ID.
func GetGenre(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var g genreWithCount
	err := db.DB.QueryRow(r.Context(),
		`SELECT g.id, g.name, g.user_id, COUNT(bg.book_id) AS book_count
		FROM genres g
		LEFT JOIN book_genres bg ON bg.genre_id = g.id
		WHERE g.id = $1 AND g.user_id = $2
		GROUP BY g.id`, id, userID).
		Scan(&g.ID, &g.Name, &g.UserID, &g.BookCount)
	if err != nil {
		http.Error(w, "genre not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

type genreBody struct {
	Name string `json:"name"`
}

// CreateGenre creates a new genre.
func CreateGenre(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body genreBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var g models.Genre
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO genres (name, user_id) VALUES ($1, $2)
		ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, user_id`,
		body.Name, userID).Scan(&g.ID, &g.Name, &g.UserID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

// UpdateGenre renames a genre.
func UpdateGenre(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var body genreBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var g models.Genre
	err := db.DB.QueryRow(r.Context(),
		`UPDATE genres SET name = $1 WHERE id = $2 AND user_id = $3 RETURNING id, name, user_id`,
		body.Name, id, userID).Scan(&g.ID, &g.Name, &g.UserID)
	if err != nil {
		http.Error(w, "genre not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

// DeleteGenre removes a genre.
func DeleteGenre(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM genres WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "genre not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListGenreBooks returns all books in a given genre.
func ListGenreBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM genres WHERE id = $1 AND user_id = $2)",
		id, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "genre not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_genres bg ON bg.book_id = b.id
		WHERE bg.genre_id = $1
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
