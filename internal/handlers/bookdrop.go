package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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
const stagedBookCols = `id, title, subtitle, author, subject, description, publisher, contributor,
	date, type, format, identifier, source, language, relation, coverage,
	series_name, series_number, series_total, page_count, rating, tags,
	cover_image IS NOT NULL AS has_cover,
	file_name, ext, original_path, user_id`

func scanStagedBook(scan func(dest ...any) error) (models.StagedBook, error) {
	var b models.StagedBook
	err := scan(
		&b.ID, &b.Title, &b.Subtitle, &b.Author,
		&b.Subject, &b.Description, &b.Publisher, &b.Contributor,
		&b.Date, &b.Type, &b.Format, &b.Identifier,
		&b.Source, &b.Language, &b.Relation, &b.Coverage,
		&b.SeriesName, &b.SeriesNumber, &b.SeriesTotal, &b.PageCount, &b.Rating, &b.Tags,
		&b.HasCover,
		&b.FileName, &b.Ext, &b.OriginalPath, &b.UserID,
	)
	return b, err
}

// ScanBookdrop scans the bookdrop directory and inserts new files into staged_books.
func ScanBookdrop(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	if !TryLockScan("bookdrop_" + userID) {
		http.Error(w, "bookdrop scan already in progress for this user", http.StatusTooManyRequests)
		return
	}
	defer UnlockScan("bookdrop_" + userID)

	targetDir := r.URL.Query().Get("path")
	if targetDir == "" {
		// Fall back to the user's saved bookdrop path
		var saved *string
		_ = db.DB.QueryRow(r.Context(),
			"SELECT bookdrop_path FROM user_settings WHERE user_id = $1", userID).Scan(&saved)
		if saved != nil && *saved != "" {
			targetDir = *saved
		}
	}
	if targetDir == "" {
		http.Error(w, "no bookdrop path configured", http.StatusBadRequest)
		return
	}

	cleanedDir, err := ValidatePath(targetDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	info, err := os.Stat(cleanedDir)
	if err != nil || !info.IsDir() {
		http.Error(w, "bookdrop directory does not exist or is not a directory", http.StatusBadRequest)
		return
	}

	type bookFile struct {
		name         string
		originalPath string
	}
	var files []bookFile
	err = filepath.WalkDir(cleanedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if validBookExts[ext] {
			files = append(files, bookFile{name: d.Name(), originalPath: path})
		}
		return nil
	})
	if err != nil {
		http.Error(w, "failed to read bookdrop directory", http.StatusInternalServerError)
		return
	}

	// Pre-fetch existing original_paths to avoid N+1 queries
	existingPaths := make(map[string]bool)
	rows, err := db.DB.Query(r.Context(), "SELECT original_path FROM staged_books WHERE user_id = $1", userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p string
			if err := rows.Scan(&p); err == nil {
				existingPaths[p] = true
			}
		}
	}

	for _, f := range files {
		baseName := f.name
		originalPath := f.originalPath
		ext := strings.ToLower(filepath.Ext(baseName))

		if existingPaths[originalPath] {
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
	Title        *string  `json:"title"`
	Subtitle     *string  `json:"subtitle"`
	Author       *string  `json:"author"`
	Subject      *string  `json:"subject"`
	Description  *string  `json:"description"`
	Publisher    *string  `json:"publisher"`
	Contributor  *string  `json:"contributor"`
	Date         *string  `json:"date"`
	Type         *string  `json:"type"`
	Format       *string  `json:"format"`
	Identifier   *string  `json:"identifier"`
	Source       *string  `json:"source"`
	Language     *string  `json:"language"`
	Relation     *string  `json:"relation"`
	Coverage     *string  `json:"coverage"`
	SeriesName   *string  `json:"seriesName"`
	SeriesNumber *float64 `json:"seriesNumber"`
	SeriesTotal  *int     `json:"seriesTotal"`
	PageCount    *int     `json:"pageCount"`
	Rating       *int     `json:"rating"`
	Tags         *string  `json:"tags"`
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
	if body.Subtitle != nil {
		existing.Subtitle = body.Subtitle
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
	if body.SeriesName != nil {
		existing.SeriesName = body.SeriesName
	}
	if body.SeriesNumber != nil {
		existing.SeriesNumber = body.SeriesNumber
	}
	if body.SeriesTotal != nil {
		existing.SeriesTotal = body.SeriesTotal
	}
	if body.PageCount != nil {
		existing.PageCount = body.PageCount
	}
	if body.Rating != nil {
		existing.Rating = body.Rating
	}
	if body.Tags != nil {
		existing.Tags = body.Tags
	}

	_, err = db.DB.Exec(r.Context(),
		`UPDATE staged_books SET
			title=$1, subtitle=$2, author=$3, subject=$4, description=$5, publisher=$6,
			contributor=$7, date=$8, type=$9, format=$10, identifier=$11, source=$12,
			language=$13, relation=$14, coverage=$15,
			series_name=$16, series_number=$17, series_total=$18, page_count=$19, rating=$20,
			tags=$21
		WHERE id = $22 AND user_id = $23`,
		existing.Title, existing.Subtitle, existing.Author,
		existing.Subject, existing.Description, existing.Publisher, existing.Contributor,
		existing.Date, existing.Type, existing.Format, existing.Identifier,
		existing.Source, existing.Language, existing.Relation, existing.Coverage,
		existing.SeriesName, existing.SeriesNumber, existing.SeriesTotal, existing.PageCount, existing.Rating,
		existing.Tags,
		id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

type stagedBulkUpdate struct {
	IDs         []string `json:"ids"`
	SeriesName  *string  `json:"seriesName"`
	Publisher   *string  `json:"publisher"`
	Language    *string  `json:"language"`
	SeriesTotal *int     `json:"seriesTotal"`
	Authors     []string `json:"authors"`
	AuthorsMode string   `json:"authorsMode"` // "replace" or "merge"
	Genres      []string `json:"genres"`
	GenresMode  string   `json:"genresMode"`
	Tags        []string `json:"tags"`
	TagsMode    string   `json:"tagsMode"`
}

// BulkUpdateStagedBooks applies the same field(s) to multiple staged books at once.
func BulkUpdateStagedBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body stagedBulkUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	for _, id := range body.IDs {
		// For array fields with merge mode, read existing value first
		var existingAuthor, existingSubject, existingTags *string
		if (len(body.Authors) > 0 && body.AuthorsMode == "merge") ||
			(len(body.Genres) > 0 && body.GenresMode == "merge") ||
			(len(body.Tags) > 0 && body.TagsMode == "merge") {
			_ = tx.QueryRow(r.Context(),
				"SELECT author, subject, tags FROM staged_books WHERE id = $1 AND user_id = $2",
				id, userID).Scan(&existingAuthor, &existingSubject, &existingTags)
		}

		var newAuthor, newSubject, newTags *string

		if len(body.Authors) > 0 {
			if body.AuthorsMode == "merge" {
				merged := mergeCSV(existingAuthor, body.Authors)
				newAuthor = &merged
			} else {
				joined := strings.Join(body.Authors, ", ")
				newAuthor = &joined
			}
		}
		if len(body.Genres) > 0 {
			if body.GenresMode == "merge" {
				merged := mergeCSV(existingSubject, body.Genres)
				newSubject = &merged
			} else {
				joined := strings.Join(body.Genres, ", ")
				newSubject = &joined
			}
		}
		if len(body.Tags) > 0 {
			if body.TagsMode == "merge" {
				merged := mergeCSV(existingTags, body.Tags)
				newTags = &merged
			} else {
				joined := strings.Join(body.Tags, ", ")
				newTags = &joined
			}
		}

		_, err = tx.Exec(r.Context(),
			`UPDATE staged_books SET
				series_name  = COALESCE($1, series_name),
				publisher    = COALESCE($2, publisher),
				language     = COALESCE($3, language),
				series_total = COALESCE($4, series_total),
				author       = COALESCE($5, author),
				subject      = COALESCE($6, subject),
				tags         = COALESCE($7, tags)
			WHERE id = $8 AND user_id = $9`,
			body.SeriesName, body.Publisher, body.Language, body.SeriesTotal,
			newAuthor, newSubject, newTags,
			id, userID)
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

// mergeCSV merges incoming strings into an existing comma-separated list, deduplicating.
func mergeCSV(existing *string, incoming []string) string {
	seen := map[string]bool{}
	var result []string
	if existing != nil && *existing != "" {
		for _, s := range strings.Split(*existing, ",") {
			s = strings.TrimSpace(s)
			if s != "" && !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
	}
	for _, s := range incoming {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return strings.Join(result, ", ")
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
		patternData := buildPatternData(book.Title, book.Author, book.Date, book.Publisher, book.Language, book.SeriesName, book.SeriesNumber, book.Ext)
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

		// Resolution of relations should be inside the transaction or part of the atomic unit.
		tx, err := db.DB.Begin(r.Context())
		if err != nil {
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		var authors []models.Author
		if book.Author != nil && *book.Author != "" {
			authors, err = findOrCreateAuthorsTX(r, tx, parseAuthorString(*book.Author), userID)
			if err != nil {
				tx.Rollback(r.Context())
				res.Error = "failed to resolve authors"
				results = append(results, res)
				continue
			}
		}
		var genres []models.Genre
		if book.Subject != nil && *book.Subject != "" {
			genres, err = findOrCreateGenresTX(r, tx, strings.Split(*book.Subject, ","), userID)
			if err != nil {
				tx.Rollback(r.Context())
				res.Error = "failed to resolve genres"
				results = append(results, res)
				continue
			}
		}
		var tags []models.Tag
		if book.Tags != nil && *book.Tags != "" {
			tags, err = findOrCreateTagsTX(r, tx, strings.Split(*book.Tags, ","), userID)
			if err != nil {
				tx.Rollback(r.Context())
				res.Error = "failed to resolve tags"
				results = append(results, res)
				continue
			}
		}

		var bookID string
		err = tx.QueryRow(r.Context(),
			`INSERT INTO books (library_id, user_id, file_path)
			VALUES ($1, $2, $3) RETURNING id`,
			item.LibraryID, userID, destPath).Scan(&bookID)
		if err != nil {
			tx.Rollback(r.Context())
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		// Resolve cover path
		var coverImage []byte
		var coverMimeStr *string
		_ = tx.QueryRow(r.Context(),
			"SELECT cover_image, cover_mime FROM staged_books WHERE id = $1",
			item.StagedBookID).Scan(&coverImage, &coverMimeStr)

		var coverPath *string
		var coverMimeVal *string
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

		_, err = tx.Exec(r.Context(),
			`INSERT INTO book_metadata (
				book_id, title, subtitle, description, publisher, published_date, language,
				series_name, series_number, series_total, page_count, rating,
				cover_path, cover_mime
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			bookID, book.Title, book.Subtitle,
			book.Description, book.Publisher, book.Date, book.Language,
			book.SeriesName, book.SeriesNumber, book.SeriesTotal, book.PageCount, book.Rating,
			coverPath, coverMimeVal)
		if err != nil {
			tx.Rollback(r.Context())
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		_ = linkBookAuthors(r, tx, bookID, authors)
		_ = linkBookGenres(r, tx, bookID, genres)
		_ = linkBookTags(r, tx, bookID, tags)

		_, err = tx.Exec(r.Context(),
			"DELETE FROM staged_books WHERE id = $1 AND user_id = $2",
			item.StagedBookID, userID)
		if err != nil {
			tx.Rollback(r.Context())
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		if err := tx.Commit(r.Context()); err != nil {
			res.Error = fmt.Sprintf("db error: %v", err)
			results = append(results, res)
			continue
		}

		results = append(results, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
