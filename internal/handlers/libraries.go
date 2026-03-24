package handlers

import (
	"encoding/json"
	"fmt"
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

type libraryBody struct {
	Name   string  `json:"name"`
	Icon   *string `json:"icon"`
	Folder *string `json:"folder"`
}

func ListLibraries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, err := db.DB.Query(r.Context(), `
		SELECT l.id, l.name, l.icon, l.folder, l.user_id, COUNT(b.id) AS book_count
		FROM libraries l
		LEFT JOIN books b ON b.library_id = l.id
		WHERE l.user_id = $1
		GROUP BY l.id, l.name, l.icon, l.folder, l.user_id`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	libraries := []models.Library{}
	for rows.Next() {
		var l models.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Icon, &l.Folder, &l.UserID, &l.BookCount); err != nil {
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

type scanResult struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
}

// ScanLibrary scans a single library's folder, adds new books and removes missing ones.
func ScanLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var lib models.Library
	err := db.DB.QueryRow(r.Context(),
		"SELECT id, name, icon, folder, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&lib.ID, &lib.Name, &lib.Icon, &lib.Folder, &lib.UserID)
	if err != nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	result, err := scanLibraryFolder(r, lib.ID, *lib.Folder, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ScanAllLibraries scans all libraries for the authenticated user.
func ScanAllLibraries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := db.DB.Query(r.Context(),
		"SELECT id, folder FROM libraries WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	total := scanResult{}
	for rows.Next() {
		var id, folder string
		if err := rows.Scan(&id, &folder); err != nil {
			continue
		}
		result, err := scanLibraryFolder(r, id, folder, userID)
		if err != nil {
			continue
		}
		total.Added += result.Added
		total.Removed += result.Removed
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(total)
}

func scanLibraryFolder(r *http.Request, libraryID, folder, userID string) (scanResult, error) {
	var result scanResult

	// 1. Collect all book file paths currently on disk
	diskFiles := map[string]bool{}
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if validBookExts[ext] {
			diskFiles[path] = true
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("failed to walk library folder: %v", err)
	}

	// 2. Get all book file_paths in DB for this library
	rows, err := db.DB.Query(r.Context(),
		"SELECT id, file_path FROM books WHERE library_id = $1 AND user_id = $2", libraryID, userID)
	if err != nil {
		return result, fmt.Errorf("db error: %v", err)
	}
	defer rows.Close()

	dbFiles := map[string]string{} // file_path -> book id
	for rows.Next() {
		var id, fp string
		if err := rows.Scan(&id, &fp); err != nil {
			continue
		}
		dbFiles[fp] = id
	}

	// 3. Remove books whose files no longer exist on disk
	for fp, id := range dbFiles {
		if !diskFiles[fp] {
			_, _ = db.DB.Exec(r.Context(),
				"DELETE FROM books WHERE id = $1 AND user_id = $2", id, userID)
			result.Removed++
		}
	}

	// 4. Add new files not yet in DB
	for fp := range diskFiles {
		if _, exists := dbFiles[fp]; exists {
			continue
		}

		meta := metadata.Extract(fp)
		title := meta.Title
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
		}

		var coverImage []byte
		var coverMime *string
		if len(meta.CoverImage) > 0 {
			coverImage = meta.CoverImage
			coverMime = &meta.CoverMime
		}

		var bookID string
		err := db.DB.QueryRow(r.Context(),
			`INSERT INTO books (
				library_id, user_id, title, subject, description, publisher, contributor,
				date, type, format, identifier, source, language, relation, coverage,
				cover_image, cover_mime, file_path
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			RETURNING id`,
			libraryID, userID,
			title,
			nilIfEmpty(meta.Subject), nilIfEmpty(meta.Description),
			nilIfEmpty(meta.Publisher), nilIfEmpty(meta.Contributor),
			nilIfEmpty(meta.Date), nilIfEmpty(meta.Type),
			nilIfEmpty(meta.Format), nilIfEmpty(meta.Identifier),
			nilIfEmpty(meta.Source), nilIfEmpty(meta.Language),
			nilIfEmpty(meta.Relation), nilIfEmpty(meta.Coverage),
			coverImage, coverMime, fp).Scan(&bookID)
		if err == nil {
			result.Added++
			// Link authors from metadata
			if meta.Creator != "" {
				authorNames := parseAuthorString(meta.Creator)
				authors, err := findOrCreateAuthors(r, authorNames, userID)
				if err == nil {
					_ = linkBookAuthors(r, bookID, authors)
				}
			}
		}
	}

	return result, nil
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
