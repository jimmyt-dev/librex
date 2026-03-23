package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
	rows, err := db.DB.QueryContext(r.Context(), "SELECT id, name, icon, user_id FROM shelves WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	shelves := []models.Shelf{}
	for rows.Next() {
		var s models.Shelf
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon, &s.UserID); err != nil {
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
	err := db.DB.QueryRowContext(r.Context(), "SELECT id, name, icon, user_id FROM shelves WHERE id = ? AND user_id = ?", id, userID).
		Scan(&s.ID, &s.Name, &s.Icon, &s.UserID)
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

	id := uuid.New().String()
	_, err := db.DB.ExecContext(r.Context(), "INSERT INTO shelves (id, name, icon, user_id) VALUES (?, ?, ?, ?)", id, body.Name, body.Icon, userID)
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
	err := db.DB.QueryRowContext(r.Context(), "SELECT id, name, icon, user_id FROM shelves WHERE id = ? AND user_id = ?", id, userID).
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

	_, execErr := db.DB.ExecContext(r.Context(), "UPDATE shelves SET name = ?, icon = ? WHERE id = ? AND user_id = ?", body.Name, body.Icon, id, userID)
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

	result, err := db.DB.ExecContext(r.Context(), "DELETE FROM shelves WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "shelf not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
