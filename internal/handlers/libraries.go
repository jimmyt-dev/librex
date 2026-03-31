package handlers

import (
	"encoding/json"
	"fmt"
	"io"
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
	Name              string  `json:"name"`
	Icon              *string `json:"icon"`
	Folder            *string `json:"folder"`
	FileNamingPattern *string `json:"fileNamingPattern"`
}

func ListLibraries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	rows, err := db.DB.Query(r.Context(), `
		SELECT l.id, l.name, l.icon, l.folder, l.file_naming_pattern, l.user_id, COUNT(b.id) AS book_count
		FROM libraries l
		LEFT JOIN books b ON b.library_id = l.id
		WHERE l.user_id = $1
		GROUP BY l.id, l.name, l.icon, l.folder, l.file_naming_pattern, l.user_id`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	libraries := []models.Library{}
	for rows.Next() {
		var l models.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Icon, &l.Folder, &l.FileNamingPattern, &l.UserID, &l.BookCount); err != nil {
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
	err := db.DB.QueryRow(r.Context(), "SELECT id, name, icon, folder, file_naming_pattern, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&l.ID, &l.Name, &l.Icon, &l.Folder, &l.FileNamingPattern, &l.UserID)
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
		"INSERT INTO libraries (name, icon, folder, file_naming_pattern, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		body.Name, body.Icon, body.Folder, body.FileNamingPattern, userID).Scan(&id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Library{ID: id, Name: body.Name, Icon: body.Icon, Folder: body.Folder, FileNamingPattern: body.FileNamingPattern, UserID: userID})
}

func UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var existing models.Library
	err := db.DB.QueryRow(r.Context(), "SELECT id, name, icon, folder, file_naming_pattern, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&existing.ID, &existing.Name, &existing.Icon, &existing.Folder, &existing.FileNamingPattern, &existing.UserID)
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
		"UPDATE libraries SET name = $1, icon = $2, folder = $3, file_naming_pattern = $4 WHERE id = $5 AND user_id = $6",
		body.Name, body.Icon, body.Folder, body.FileNamingPattern, id, userID)
	if execErr != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Library{ID: id, Name: body.Name, Icon: body.Icon, Folder: body.Folder, FileNamingPattern: body.FileNamingPattern, UserID: userID})
}

func DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	deleteFiles := r.URL.Query().Get("deleteFiles") == "true"

	var folder *string
	_ = db.DB.QueryRow(r.Context(),
		"SELECT folder FROM libraries WHERE id = $1 AND user_id = $2", id, userID).Scan(&folder)

	// If deleting files, fetch all file paths first
	var filePaths []string
	if deleteFiles {
		rows, err := db.DB.Query(r.Context(), "SELECT file_path FROM books WHERE library_id = $1 AND user_id = $2", id, userID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var fp string
				if err := rows.Scan(&fp); err == nil {
					filePaths = append(filePaths, fp)
				}
			}
		}
	}

	result, err := db.DB.Exec(r.Context(), "DELETE FROM libraries WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	// Cleanup covers
	if folder != nil && *folder != "" {
		os.RemoveAll(filepath.Join(*folder, ".covers"))
	}

	// Cleanup book files if requested
	if deleteFiles {
		for _, fp := range filePaths {
			os.Remove(fp)
		}
	}

	// Cleanup orphaned metadata entities
	_ = CleanupOrphanAuthors(r, db.DB, userID)
	_ = CleanupOrphanGenres(r, db.DB, userID)
	_ = CleanupOrphanTags(r, db.DB, userID)

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

	if !TryLockScan("lib_" + id) {
		http.Error(w, "scan already in progress for this library", http.StatusTooManyRequests)
		return
	}
	defer UnlockScan("lib_" + id)

	var lib models.Library
	err := db.DB.QueryRow(r.Context(),
		"SELECT id, name, icon, folder, file_naming_pattern, user_id FROM libraries WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&lib.ID, &lib.Name, &lib.Icon, &lib.Folder, &lib.FileNamingPattern, &lib.UserID)
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
			var coverPath *string
			_ = db.DB.QueryRow(r.Context(),
				"SELECT cover_path FROM book_metadata WHERE book_id = $1", id).Scan(&coverPath)
			_, _ = db.DB.Exec(r.Context(),
				"DELETE FROM books WHERE id = $1 AND user_id = $2", id, userID)
			if coverPath != nil && *coverPath != "" {
				os.Remove(*coverPath)
			}
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

		// Use a transaction per book for atomicity
		tx, err := db.DB.Begin(r.Context())
		if err != nil {
			continue
		}

		// Insert slim books row
		var bookID string
		err = tx.QueryRow(r.Context(),
			`INSERT INTO books (library_id, user_id, file_path)
			VALUES ($1, $2, $3) RETURNING id`,
			libraryID, userID, fp).Scan(&bookID)
		if err != nil {
			tx.Rollback(r.Context())
			continue
		}

		// Write cover to disk
		var coverPath *string
		var coverMimeVal *string
		if len(meta.CoverImage) > 0 {
			if cp, err := writeCoverToDisk(folder, bookID, meta.CoverImage, meta.CoverMime); err == nil {
				coverPath = &cp
				coverMimeVal = &meta.CoverMime
			}
		}

		// Insert book_metadata row
		_, metaErr := tx.Exec(r.Context(),
			`INSERT INTO book_metadata (
				book_id, title, description, publisher, published_date, language,
				cover_path, cover_mime
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			bookID, title,
			nilIfEmpty(meta.Description), nilIfEmpty(meta.Publisher),
			nilIfEmpty(meta.Date), nilIfEmpty(meta.Language),
			coverPath, coverMimeVal)
		if metaErr != nil {
			tx.Rollback(r.Context())
			continue
		}

		// Link authors from metadata
		if meta.Creator != "" {
			authorNames := parseAuthorString(meta.Creator)
			authors, err := findOrCreateAuthorsTX(r, tx, authorNames, userID)
			if err == nil {
				_ = linkBookAuthors(r, tx, bookID, authors)
			}
		}

		// Link genres from subject
		if meta.Subject != "" {
			genreNames := strings.Split(meta.Subject, ",")
			genres, err := findOrCreateGenresTX(r, tx, genreNames, userID)
			if err == nil {
				_ = linkBookGenres(r, tx, bookID, genres)
			}
		}

		if err := tx.Commit(r.Context()); err != nil {
			continue
		}

		result.Added++
	}

	return result, nil
}

// UploadToLibrary accepts multipart file uploads and adds them directly to a library.
// POST /api/libraries/{id}/upload
func UploadToLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	libraryID := chi.URLParam(r, "id")

	var lib models.Library
	err := db.DB.QueryRow(r.Context(),
		"SELECT id, name, icon, folder, file_naming_pattern, user_id FROM libraries WHERE id = $1 AND user_id = $2",
		libraryID, userID).
		Scan(&lib.ID, &lib.Name, &lib.Icon, &lib.Folder, &lib.FileNamingPattern, &lib.UserID)
	if err != nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}
	if lib.Folder == nil || *lib.Folder == "" {
		http.Error(w, "library has no folder configured", http.StatusBadRequest)
		return
	}

	cleanedDir, err := ValidatePath(*lib.Folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err := os.MkdirAll(cleanedDir, 0755); err != nil {
		http.Error(w, "failed to create library directory", http.StatusInternalServerError)
		return
	}

	userID = middleware.GetUserID(r)
	maxSize := int64(500 << 20) // Default 500 MB
	var userMaxMB int
	err = db.DB.QueryRow(r.Context(), "SELECT max_upload_size_mb FROM user_settings WHERE user_id = $1", userID).Scan(&userMaxMB)
	if err == nil && userMaxMB > 0 {
		maxSize = int64(userMaxMB) << 20
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, fmt.Sprintf("Upload too large. The current limit is %d MB. You can increase this in Settings.", userMaxMB), http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "failed to parse upload: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}

	var added []models.Book

	for _, fh := range files {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !validBookExts[ext] {
			continue
		}

		src, err := fh.Open()
		if err != nil {
			continue
		}

		destPath := filepath.Join(cleanedDir, filepath.Base(fh.Filename))
		if _, err := os.Stat(destPath); err == nil {
			base := strings.TrimSuffix(filepath.Base(fh.Filename), ext)
			destPath = filepath.Join(cleanedDir, fmt.Sprintf("%s_%d%s", base, os.Getpid(), ext))
		}

		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			continue
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if copyErr != nil {
			os.Remove(destPath)
			continue
		}

		// Skip if already in the library
		var alreadyExists bool
		_ = db.DB.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM books WHERE file_path = $1 AND user_id = $2)",
			destPath, userID).Scan(&alreadyExists)
		if alreadyExists {
			continue
		}

		meta := metadata.Extract(destPath)
		title := meta.Title
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(fh.Filename), ext)
		}

		tx, err := db.DB.Begin(r.Context())
		if err != nil {
			continue
		}

		var bookID string
		err = tx.QueryRow(r.Context(),
			`INSERT INTO books (library_id, user_id, file_path) VALUES ($1, $2, $3) RETURNING id`,
			libraryID, userID, destPath).Scan(&bookID)
		if err != nil {
			tx.Rollback(r.Context())
			continue
		}

		var coverPath *string
		var coverMimeVal *string
		if len(meta.CoverImage) > 0 {
			if cp, err := writeCoverToDisk(cleanedDir, bookID, meta.CoverImage, meta.CoverMime); err == nil {
				coverPath = &cp
				coverMimeVal = &meta.CoverMime
			}
		}

		_, metaErr := tx.Exec(r.Context(),
			`INSERT INTO book_metadata (
				book_id, title, description, publisher, published_date, language,
				cover_path, cover_mime
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			bookID, title,
			nilIfEmpty(meta.Description), nilIfEmpty(meta.Publisher),
			nilIfEmpty(meta.Date), nilIfEmpty(meta.Language),
			coverPath, coverMimeVal)
		if metaErr != nil {
			tx.Rollback(r.Context())
			continue
		}

		if meta.Creator != "" {
			authorNames := parseAuthorString(meta.Creator)
			authors, err := findOrCreateAuthorsTX(r, tx, authorNames, userID)
			if err == nil {
				_ = linkBookAuthors(r, tx, bookID, authors)
			}
		}

		if meta.Subject != "" {
			genreNames := strings.Split(meta.Subject, ",")
			genres, err := findOrCreateGenresTX(r, tx, genreNames, userID)
			if err == nil {
				_ = linkBookGenres(r, tx, bookID, genres)
			}
		}

		if err := tx.Commit(r.Context()); err != nil {
			continue
		}

		book, fetchErr := fetchBookByID(r, bookID, userID)
		if fetchErr == nil {
			added = append(added, book)
		}
	}

	if added == nil {
		added = []models.Book{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(added)
}

func validateFolder(r *http.Request, folder *string, excludeID string) *validationError {
	if folder == nil || *folder == "" {
		return &validationError{"folder cannot be null or empty", http.StatusBadRequest}
	}

	cleaned, err := ValidatePath(*folder)
	if err != nil {
		return &validationError{err.Error(), http.StatusForbidden}
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
