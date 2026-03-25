package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/metadata"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

var validBookExts = map[string]bool{
	".epub": true,
	".pdf":  true,
	".mobi": true,
	".azw3": true,
	".cbz":  true,
	".cbr":  true,
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// stagedBookCols selects all fields for list/get responses.
// cover_image is excluded; cover_image IS NOT NULL is used to set has_cover.
const stagedBookCols = `id, title, author, subject, description, publisher, contributor,
	date, type, format, identifier, source, language, relation, coverage,
	cover_image IS NOT NULL AS has_cover,
	file_name, ext, original_path, user_id`

func scanStagedBook(scan func(dest ...any) error) (models.StagedBook, error) {
	var b models.StagedBook
	err := scan(
		&b.ID, &b.Title, &b.Author,
		&b.Subject, &b.Description, &b.Publisher, &b.Contributor,
		&b.Date, &b.Type, &b.Format, &b.Identifier,
		&b.Source, &b.Language, &b.Relation, &b.Coverage,
		&b.HasCover,
		&b.FileName, &b.Ext, &b.OriginalPath, &b.UserID,
	)
	return b, err
}

// ScanBookdrop scans the bookdrop directory and inserts new files into staged_books.
func ScanBookdrop(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	targetDir := r.URL.Query().Get("path")
	if targetDir == "" {
		targetDir = "/Users/jimmy/Documents/Code Shit/reliquary/data/bookdrop"
	}

	cleanedDir := filepath.Clean(targetDir)

	if !filepath.IsAbs(cleanedDir) {
		http.Error(w, "bookdrop path must be absolute", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(cleanedDir)
	if err != nil || !info.IsDir() {
		http.Error(w, "bookdrop directory does not exist or is not a directory", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(cleanedDir)
	if err != nil {
		http.Error(w, "failed to read bookdrop directory", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !validBookExts[ext] {
			continue
		}

		baseName := entry.Name()
		originalPath := filepath.Join(cleanedDir, baseName)

		var count int
		if err := db.DB.QueryRow(r.Context(),
			"SELECT COUNT(*) FROM staged_books WHERE original_path = $1 AND user_id = $2",
			originalPath, userID).Scan(&count); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if count > 0 {
			continue
		}

		meta := metadata.Extract(originalPath)
		title := meta.Title
		if title == "" {
			title = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		}

		var coverImage []byte
		var coverMime *string
		if len(meta.CoverImage) > 0 {
			coverImage = meta.CoverImage
			coverMime = &meta.CoverMime
		}

		_, err = db.DB.Exec(r.Context(),
			`INSERT INTO staged_books (
				title, author, subject, description, publisher, contributor,
				date, type, format, identifier, source, language, relation, coverage,
				cover_image, cover_mime,
				file_name, ext, original_path, user_id
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
			title, nilIfEmpty(meta.Creator),
			nilIfEmpty(meta.Subject), nilIfEmpty(meta.Description),
			nilIfEmpty(meta.Publisher), nilIfEmpty(meta.Contributor),
			nilIfEmpty(meta.Date), nilIfEmpty(meta.Type),
			nilIfEmpty(meta.Format), nilIfEmpty(meta.Identifier),
			nilIfEmpty(meta.Source), nilIfEmpty(meta.Language),
			nilIfEmpty(meta.Relation), nilIfEmpty(meta.Coverage),
			coverImage, coverMime,
			baseName, ext, originalPath, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	listStagedBooks(w, r, userID)
}

// ListStagedBooks returns all staged books for the authenticated user.
func ListStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	listStagedBooks(w, r, userID)
}

func listStagedBooks(w http.ResponseWriter, r *http.Request, userID string) {
	rows, err := db.DB.Query(r.Context(),
		"SELECT "+stagedBookCols+" FROM staged_books WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []models.StagedBook{}
	for rows.Next() {
		b, err := scanStagedBook(rows.Scan)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetStagedBook returns a single staged book by ID.
func GetStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		"SELECT "+stagedBookCols+" FROM staged_books WHERE id = $1 AND user_id = $2",
		id, userID)
	b, err := scanStagedBook(row.Scan)
	if err != nil {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// GetStagedBookCover serves the raw cover image for a staged book.
func GetStagedBookCover(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var imgData []byte
	var mime string
	err := db.DB.QueryRow(r.Context(),
		"SELECT cover_image, cover_mime FROM staged_books WHERE id = $1 AND user_id = $2",
		id, userID).Scan(&imgData, &mime)
	if err != nil || len(imgData) == 0 {
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	}

	if mime == "" {
		mime = "image/jpeg"
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(imgData)
}

type stagedBookUpdate struct {
	Title       *string `json:"title"`
	Author      *string `json:"author"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	Publisher   *string `json:"publisher"`
	Contributor *string `json:"contributor"`
	Date        *string `json:"date"`
	Type        *string `json:"type"`
	Format      *string `json:"format"`
	Identifier  *string `json:"identifier"`
	Source      *string `json:"source"`
	Language    *string `json:"language"`
	Relation    *string `json:"relation"`
	Coverage    *string `json:"coverage"`
}

// UpdateStagedBook updates the editable metadata of a staged book.
func UpdateStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		"SELECT "+stagedBookCols+" FROM staged_books WHERE id = $1 AND user_id = $2",
		id, userID)
	existing, err := scanStagedBook(row.Scan)
	if err != nil {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	var body stagedBookUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Title != nil {
		existing.Title = *body.Title
	}
	if body.Author != nil {
		existing.Author = body.Author
	}
	if body.Subject != nil {
		existing.Subject = body.Subject
	}
	if body.Description != nil {
		existing.Description = body.Description
	}
	if body.Publisher != nil {
		existing.Publisher = body.Publisher
	}
	if body.Contributor != nil {
		existing.Contributor = body.Contributor
	}
	if body.Date != nil {
		existing.Date = body.Date
	}
	if body.Type != nil {
		existing.Type = body.Type
	}
	if body.Format != nil {
		existing.Format = body.Format
	}
	if body.Identifier != nil {
		existing.Identifier = body.Identifier
	}
	if body.Source != nil {
		existing.Source = body.Source
	}
	if body.Language != nil {
		existing.Language = body.Language
	}
	if body.Relation != nil {
		existing.Relation = body.Relation
	}
	if body.Coverage != nil {
		existing.Coverage = body.Coverage
	}

	_, err = db.DB.Exec(r.Context(),
		`UPDATE staged_books SET
			title=$1, author=$2, subject=$3, description=$4, publisher=$5, contributor=$6,
			date=$7, type=$8, format=$9, identifier=$10, source=$11, language=$12,
			relation=$13, coverage=$14
		WHERE id = $15 AND user_id = $16`,
		existing.Title, existing.Author,
		existing.Subject, existing.Description, existing.Publisher, existing.Contributor,
		existing.Date, existing.Type, existing.Format, existing.Identifier,
		existing.Source, existing.Language, existing.Relation, existing.Coverage,
		id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

type bulkUpdateItem struct {
	ID     string  `json:"id"`
	Title  *string `json:"title"`
	Author *string `json:"author"`
}

// BulkUpdateStagedBooks updates multiple staged books at once.
func BulkUpdateStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var items []bulkUpdateItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	for _, item := range items {
		if item.Title != nil && item.Author != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET title = $1, author = $2 WHERE id = $3 AND user_id = $4",
				*item.Title, *item.Author, item.ID, userID)
		} else if item.Title != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET title = $1 WHERE id = $2 AND user_id = $3",
				*item.Title, item.ID, userID)
		} else if item.Author != nil {
			_, err = tx.Exec(r.Context(),
				"UPDATE staged_books SET author = $1 WHERE id = $2 AND user_id = $3",
				*item.Author, item.ID, userID)
		}
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	listStagedBooks(w, r, userID)
}

// DeleteStagedBook removes a staged book by ID.
func DeleteStagedBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM staged_books WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "staged book not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearStagedBooks removes all staged books for the authenticated user.
func ClearStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	_, err := db.DB.Exec(r.Context(),
		"DELETE FROM staged_books WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Import ---

type importItem struct {
	StagedBookID string `json:"stagedBookId"`
	LibraryID    string `json:"libraryId"`
}

type importResult struct {
	StagedBookID string `json:"stagedBookId"`
	Error        string `json:"error,omitempty"`
}

var unsafeChars = regexp.MustCompile(`[/\\:*?"<>|]`)

// sanitizeName makes a string safe to use as a filesystem directory/file component.
func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = unsafeChars.ReplaceAllString(s, "_")
	s = strings.TrimSpace(s)
	if s == "" {
		return "Unknown"
	}
	return s
}

// moveFile moves src to dst, falling back to copy+delete on cross-device links.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Cross-device or other rename failure — copy then delete.
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		os.Remove(dst)
		return err
	}
	if err = out.Close(); err != nil {
		os.Remove(dst)
		return err
	}
	return os.Remove(src)
}

// uniqueDest returns dst unchanged if it doesn't exist, otherwise appends " (n)" before the extension.
func uniqueDest(dst string) string {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return dst
	}
	ext := filepath.Ext(dst)
	base := strings.TrimSuffix(dst, ext)
	for n := 2; n < 1000; n++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, n, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return dst
}

// ImportBooks moves staged books into their assigned libraries.
// POST /api/bookdrop/import
// Body: [{stagedBookId, libraryId}]
// Returns: [{stagedBookId, error?}] — one entry per input item.
func ImportBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var items []importItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil || len(items) == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	results := make([]importResult, 0, len(items))

	for _, item := range items {
		res := importResult{StagedBookID: item.StagedBookID}

		// Fetch staged book
		row := db.DB.QueryRow(r.Context(),
			"SELECT "+stagedBookCols+" FROM staged_books WHERE id = $1 AND user_id = $2",
			item.StagedBookID, userID)
		book, err := scanStagedBook(row.Scan)
		if err != nil {
			res.Error = "staged book not found"
			results = append(results, res)
			continue
		}

		// Fetch library
		var libraryFolder string
		err = db.DB.QueryRow(r.Context(),
			"SELECT folder FROM libraries WHERE id = $1 AND user_id = $2",
			item.LibraryID, userID).Scan(&libraryFolder)
		if err != nil {
			res.Error = "library not found"
			results = append(results, res)
			continue
		}

		// Build destination path using file naming pattern
		pattern := getEffectivePattern(r, item.LibraryID, userID)
		patternData := buildPatternData(book.Title, book.Author, book.Date, book.Publisher, book.Language, book.Ext)
		relativePath := resolveFilePattern(pattern, patternData)
		destPath := uniqueDest(filepath.Join(libraryFolder, relativePath))

		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			res.Error = fmt.Sprintf("failed to create directory: %v", err)
			results = append(results, res)
			continue
		}

		if err := moveFile(book.OriginalPath, destPath); err != nil {
			res.Error = fmt.Sprintf("failed to move file: %v", err)
			results = append(results, res)
			continue
		}

		// Insert slim books row
		var bookID string
		err = db.DB.QueryRow(r.Context(),
			`INSERT INTO books (library_id, user_id, file_path)
			VALUES ($1, $2, $3) RETURNING id`,
			item.LibraryID, userID, destPath).Scan(&bookID)
		if err != nil {
			_ = moveFile(destPath, book.OriginalPath)
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		// Write cover to disk
		var coverPath *string
		var coverMimeVal *string
		var coverImage []byte
		var coverMimeStr *string
		_ = db.DB.QueryRow(r.Context(),
			"SELECT cover_image, cover_mime FROM staged_books WHERE id = $1",
			item.StagedBookID).Scan(&coverImage, &coverMimeStr)
		if len(coverImage) > 0 {
			mime := "image/jpeg"
			if coverMimeStr != nil && *coverMimeStr != "" {
				mime = *coverMimeStr
			}
			if cp, err := writeCoverToDisk(libraryFolder, bookID, coverImage, mime); err == nil {
				coverPath = &cp
				coverMimeVal = &mime
			}
		}

		// Insert book_metadata row
		_, err = db.DB.Exec(r.Context(),
			`INSERT INTO book_metadata (
				book_id, title, description, publisher, published_date, language,
				cover_path, cover_mime
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			bookID, book.Title,
			book.Description, book.Publisher, book.Date, book.Language,
			coverPath, coverMimeVal)
		if err != nil {
			// Clean up the books row
			_, _ = db.DB.Exec(r.Context(), "DELETE FROM books WHERE id = $1", bookID)
			_ = moveFile(destPath, book.OriginalPath)
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		// Link authors
		if book.Author != nil && *book.Author != "" {
			authorNames := parseAuthorString(*book.Author)
			authors, err := findOrCreateAuthors(r, authorNames, userID)
			if err == nil {
				_ = linkBookAuthors(r, bookID, authors)
			}
		}

		// Link categories from subject
		if book.Subject != nil && *book.Subject != "" {
			catNames := strings.Split(*book.Subject, ",")
			cats, err := findOrCreateCategories(r, catNames, userID)
			if err == nil {
				_ = linkBookCategories(r, bookID, cats)
			}
		}

		// Remove from staged_books
		_, _ = db.DB.Exec(r.Context(),
			"DELETE FROM staged_books WHERE id = $1 AND user_id = $2",
			item.StagedBookID, userID)

		results = append(results, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
