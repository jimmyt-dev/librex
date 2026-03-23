package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

type libraryBody struct {
	Name   string  `json:"name"`
	Icon   *string `json:"icon"`
	Folder *string `json:"folder"`
}

func ListLibraries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, err := db.DB.Query(r.Context(), "SELECT id, name, icon, folder, user_id FROM libraries WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	libraries := []models.Library{}
	for rows.Next() {
		var l models.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Icon, &l.Folder, &l.UserID); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		libraries = append(libraries, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(libraries)
}

func GetLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var l models.Library
	err := db.DB.QueryRow(r.Context(), "SELECT id, name, icon, folder, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&l.ID, &l.Name, &l.Icon, &l.Folder, &l.UserID)
	if err != nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func CreateLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body libraryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := validateFolder(r, body.Folder, ""); err != nil {
		http.Error(w, err.msg, err.code)
		return
	}
	if body.Folder != nil {
		cleaned := filepath.Clean(*body.Folder)
		body.Folder = &cleaned
	}

	var id string
	err := db.DB.QueryRow(r.Context(),
		"INSERT INTO libraries (name, icon, folder, user_id) VALUES ($1, $2, $3, $4) RETURNING id",
		body.Name, body.Icon, body.Folder, userID).Scan(&id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Library{ID: id, Name: body.Name, Icon: body.Icon, Folder: body.Folder, UserID: userID})
}

func UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var existing models.Library
	err := db.DB.QueryRow(r.Context(), "SELECT id, name, icon, folder, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&existing.ID, &existing.Name, &existing.Icon, &existing.Folder, &existing.UserID)
	if err != nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	var body libraryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := validateFolder(r, body.Folder, id); err != nil {
		http.Error(w, err.msg, err.code)
		return
	}
	if body.Folder != nil {
		cleaned := filepath.Clean(*body.Folder)
		body.Folder = &cleaned
	}

	_, execErr := db.DB.Exec(r.Context(),
		"UPDATE libraries SET name = $1, icon = $2, folder = $3 WHERE id = $4 AND user_id = $5",
		body.Name, body.Icon, body.Folder, id, userID)
	if execErr != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Library{ID: id, Name: body.Name, Icon: body.Icon, Folder: body.Folder, UserID: userID})
}

func DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(), "DELETE FROM libraries WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type validationError struct {
	msg  string
	code int
}

func validateFolder(r *http.Request, folder *string, excludeID string) *validationError {
	if folder == nil || *folder == "" {
		return &validationError{"folder cannot be null or empty", http.StatusBadRequest}
	}

	cleaned := filepath.Clean(*folder)
	if !filepath.IsAbs(cleaned) {
		return &validationError{"folder must be an absolute path", http.StatusBadRequest}
	}

	info, err := os.Stat(cleaned)
	if err != nil || !info.IsDir() {
		return &validationError{"folder does not exist or is not a directory", http.StatusBadRequest}
	}

	query := "SELECT COUNT(*) FROM libraries WHERE folder = $1"
	args := []any{cleaned}
	if excludeID != "" {
		query += fmt.Sprintf(" AND id != $%d", len(args)+1)
		args = append(args, excludeID)
	}

	var count int
	if err := db.DB.QueryRow(r.Context(), query, args...).Scan(&count); err != nil {
		return &validationError{"db error", http.StatusInternalServerError}
	}
	if count > 0 {
		return &validationError{"Folder is already used by another library", http.StatusConflict}
	}

	return nil
}
