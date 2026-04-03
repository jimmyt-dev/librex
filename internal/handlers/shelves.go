package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"librex/internal/db"
	"librex/internal/middleware"
	"librex/internal/models"
)

type shelfBody struct {
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

func ListShelves(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// Get real shelves
	rows, err := db.DB.Query(r.Context(),
		`SELECT s.id, s.name, s.icon, s.user_id, COUNT(bs.book_id) AS books
		 FROM shelves s
		 LEFT JOIN book_shelves bs ON bs.shelf_id = s.id
		 WHERE s.user_id = $1
		 GROUP BY s.id
		 ORDER BY s.name`, userID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to query shelves", nil)
		return
	}
	defer rows.Close()

	shelves := []models.Shelf{}
	for rows.Next() {
		var s models.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon, &s.UserID, &s.Books); err != nil {
			SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to scan shelf", nil)
			return
		}
		shelves = append(shelves, s)
	}

	// Add "Unshelved" virtual shelf
	var unshelvedCount int
	err = db.DB.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM books b
		 WHERE b.user_id = $1 AND NOT EXISTS (SELECT 1 FROM book_shelves bs WHERE bs.book_id = b.id)`,
		userID).Scan(&unshelvedCount)
	if err == nil {
		icon := "inbox"
		shelves = append([]models.Shelf{{
			ID:     "unshelved",
			Name:   "Unshelved",
			Icon:   &icon,
			UserID: userID,
			Books:  unshelvedCount,
		}}, shelves...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shelves)
}

func GetShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	if id == "unshelved" {
		var unshelvedCount int
		err := db.DB.QueryRow(r.Context(),
			`SELECT COUNT(*) FROM books b
			 WHERE b.user_id = $1 AND NOT EXISTS (SELECT 1 FROM book_shelves bs WHERE bs.book_id = b.id)`,
			userID).Scan(&unshelvedCount)
		if err != nil {
			SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to count unshelved books", nil)
			return
		}
		icon := "inbox"
		json.NewEncoder(w).Encode(models.Shelf{
			ID:     "unshelved",
			Name:   "Unshelved",
			Icon:   &icon,
			UserID: userID,
			Books:  unshelvedCount,
		})
		return
	}

	var s models.Shelf
	err := db.DB.QueryRow(r.Context(),
		`SELECT s.id, s.name, s.icon, s.user_id, COUNT(bs.book_id) AS books
		 FROM shelves s
		 LEFT JOIN book_shelves bs ON bs.shelf_id = s.id
		 WHERE s.id = $1 AND s.user_id = $2
		 GROUP BY s.id`, id, userID).
		Scan(&s.ID, &s.Name, &s.Icon, &s.UserID, &s.Books)
	if err != nil {
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func CreateShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body shelfBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Name is required", map[string]string{"name": "required"})
		return
	}

	var id string
	err := db.DB.QueryRow(r.Context(),
		"INSERT INTO shelves (name, icon, user_id) VALUES ($1, $2, $3) RETURNING id",
		body.Name, body.Icon, userID).Scan(&id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to create shelf", nil)
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
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
		return
	}

	var body shelfBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Name is required", map[string]string{"name": "required"})
		return
	}

	_, execErr := db.DB.Exec(r.Context(), "UPDATE shelves SET name = $1, icon = $2 WHERE id = $3 AND user_id = $4", body.Name, body.Icon, id, userID)
	if execErr != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to update shelf", nil)
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
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to delete shelf", nil)
		return
	}

	if result.RowsAffected() == 0 {
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
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
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` JOIN book_shelves bs ON bs.book_id = b.id WHERE bs.shelf_id = $1 ORDER BY m.title`,
		shelfID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to query books for shelf", nil)
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to scan book", nil)
			return
		}
		books = append(books, b)
	}

	if err := attachBookRelations(r, books); err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to attach book relations", nil)
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
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to query unshelved books", nil)
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to scan book", nil)
			return
		}
		books = append(books, b)
	}

	if err := attachBookRelations(r, books); err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to attach book relations", nil)
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

	if shelfID == "unshelved" {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot add books directly to the unshelved shelf", nil)
		return
	}

	var body shelfBooksBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "bookIds is required", map[string]string{"bookIds": "required"})
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to start transaction", nil)
		return
	}
	defer tx.Rollback(r.Context())

	// Verify shelf ownership
	var exists bool
	if err := tx.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM shelves WHERE id = $1 AND user_id = $2)",
		shelfID, userID).Scan(&exists); err != nil || !exists {
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
		return
	}

	// Verify book ownership for all provided IDs
	var count int
	err = tx.QueryRow(r.Context(),
		"SELECT COUNT(*) FROM books WHERE id = ANY($1) AND user_id = $2",
		body.BookIDs, userID).Scan(&count)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to verify books", nil)
		return
	}
	if count != len(body.BookIDs) {
		SendError(w, http.StatusForbidden, "UNAUTHORIZED", "One or more books not found or unauthorized", nil)
		return
	}

	_, err = tx.Exec(r.Context(),
		`INSERT INTO book_shelves (book_id, shelf_id)
		 SELECT unnest($1::uuid[]), $2
		 ON CONFLICT DO NOTHING`,
		body.BookIDs, shelfID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to add books to shelf", nil)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to commit transaction", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveBooksFromShelf removes books from a shelf.
func RemoveBooksFromShelf(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	shelfID := chi.URLParam(r, "id")

	if shelfID == "unshelved" {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot remove books from the unshelved shelf directly; they must be added to a real shelf", nil)
		return
	}

	var body shelfBooksBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		SendError(w, http.StatusBadRequest, "VALIDATION_ERROR", "bookIds is required", map[string]string{"bookIds": "required"})
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to start transaction", nil)
		return
	}
	defer tx.Rollback(r.Context())

	// Verify shelf ownership
	var exists bool
	if err := tx.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM shelves WHERE id = $1 AND user_id = $2)",
		shelfID, userID).Scan(&exists); err != nil || !exists {
		SendError(w, http.StatusNotFound, "NOT_FOUND", "Shelf not found", nil)
		return
	}

	// For removal, we don't strictly *need* to check book ownership as long as the shelf is yours,
	// but it's safer and more consistent to do so.
	_, err = tx.Exec(r.Context(),
		"DELETE FROM book_shelves WHERE shelf_id = $1 AND book_id = ANY($2)",
		shelfID, body.BookIDs)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to remove books from shelf", nil)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		SendError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to commit transaction", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
