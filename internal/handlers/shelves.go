package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

type shelfBody struct {
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

func ListShelves(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, err := db.DB.Query(r.Context(),
		`SELECT s.id, s.name, s.icon, s.user_id, COUNT(bs.book_id) AS books
		 FROM shelves s
		 LEFT JOIN book_shelves bs ON bs.shelf_id = s.id
		 WHERE s.user_id = $1
		 GROUP BY s.id
		 ORDER BY s.name`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	shelves := []models.Shelf{}
	for rows.Next() {
		var s models.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon, &s.UserID, &s.Books); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		shelves = append(shelves, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shelves)
}

func GetShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var s models.Shelf
	err := db.DB.QueryRow(r.Context(),
		`SELECT s.id, s.name, s.icon, s.user_id, COUNT(bs.book_id) AS books
		 FROM shelves s
		 LEFT JOIN book_shelves bs ON bs.shelf_id = s.id
		 WHERE s.id = $1 AND s.user_id = $2
		 GROUP BY s.id`, id, userID).
		Scan(&s.ID, &s.Name, &s.Icon, &s.UserID, &s.Books)
	if err != nil {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func CreateShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body shelfBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var id string
	err := db.DB.QueryRow(r.Context(),
		"INSERT INTO shelves (name, icon, user_id) VALUES ($1, $2, $3) RETURNING id",
		body.Name, body.Icon, userID).Scan(&id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Shelf{ID: id, Name: body.Name, Icon: body.Icon, UserID: userID})
}

func UpdateShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var existing models.Shelf
	err := db.DB.QueryRow(r.Context(), "SELECT id, name, icon, user_id FROM shelves WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&existing.ID, &existing.Name, &existing.Icon, &existing.UserID)
	if err != nil {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	var body shelfBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	_, execErr := db.DB.Exec(r.Context(), "UPDATE shelves SET name = $1, icon = $2 WHERE id = $3 AND user_id = $4", body.Name, body.Icon, id, userID)
	if execErr != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Shelf{ID: id, Name: body.Name, Icon: body.Icon, UserID: userID})
}

func DeleteShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(), "DELETE FROM shelves WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListShelfBooks returns all books on a shelf.
func ListShelfBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	shelfID := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM shelves WHERE id = $1 AND user_id = $2)",
		shelfID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_shelves bs ON bs.book_id = b.id WHERE bs.shelf_id = $1 ORDER BY m.title`,
		shelfID)
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

// ListUnshelvedBooks returns books not on any shelf.
func ListUnshelvedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` WHERE b.user_id = $1 AND NOT EXISTS (SELECT 1 FROM book_shelves bs WHERE bs.book_id = b.id) ORDER BY m.title`,
		userID)
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

type shelfBooksBody struct {
	BookIDs []string `json:"bookIds"`
}

// AddBooksToShelf adds books to a shelf.
func AddBooksToShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	shelfID := chi.URLParam(r, "id")

	var body shelfBooksBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		http.Error(w, "bookIds is required", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Verify shelf ownership
	var exists bool
	if err := tx.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM shelves WHERE id = $1 AND user_id = $2)",
		shelfID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	// Verify book ownership for all provided IDs
	var count int
	err = tx.QueryRow(r.Context(),
		"SELECT COUNT(*) FROM books WHERE id = ANY($1) AND user_id = $2",
		body.BookIDs, userID).Scan(&count)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if count != len(body.BookIDs) {
		http.Error(w, "one or more books not found or unauthorized", http.StatusForbidden)
		return
	}

	_, err = tx.Exec(r.Context(),
		`INSERT INTO book_shelves (book_id, shelf_id)
		 SELECT unnest($1::uuid[]), $2
		 ON CONFLICT DO NOTHING`,
		body.BookIDs, shelfID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveBooksFromShelf removes books from a shelf.
func RemoveBooksFromShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	shelfID := chi.URLParam(r, "id")

	var body shelfBooksBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		http.Error(w, "bookIds is required", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Verify shelf ownership
	var exists bool
	if err := tx.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM shelves WHERE id = $1 AND user_id = $2)",
		shelfID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	// For removal, we don't strictly *need* to check book ownership as long as the shelf is yours,
	// but it's safer and more consistent to do so.
	_, err = tx.Exec(r.Context(),
		"DELETE FROM book_shelves WHERE shelf_id = $1 AND book_id = ANY($2)",
		shelfID, body.BookIDs)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
