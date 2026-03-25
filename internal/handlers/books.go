package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/metadata"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
	"strings"
)

const bookCols = `b.id, b.library_id, b.user_id, b.file_path, b.added_on,
	m.title, m.subtitle, m.description, m.publisher, m.published_date,
	m.isbn_13, m.isbn_10, m.language, m.page_count,
	m.series_name, m.series_number, m.cover_path, m.cover_mime`

func scanBook(scan func(dest ...any) error) (models.Book, error) {
	var b models.Book
	err := scan(
		&b.ID, &b.LibraryID, &b.UserID, &b.FilePath, &b.AddedOn,
		&b.Metadata.Title, &b.Metadata.Subtitle, &b.Metadata.Description,
		&b.Metadata.Publisher, &b.Metadata.PublishedDate,
		&b.Metadata.ISBN13, &b.Metadata.ISBN10, &b.Metadata.Language, &b.Metadata.PageCount,
		&b.Metadata.SeriesName, &b.Metadata.SeriesNumber,
		&b.Metadata.CoverPath, &b.Metadata.CoverMime,
	)
	if err == nil {
		b.Metadata.BookID = b.ID
		b.Authors = []models.Author{}
		b.Categories = []models.Category{}
		b.Tags = []models.Tag{}
	}
	return b, err
}

const bookQuery = `SELECT ` + bookCols + `
	FROM books b
	JOIN book_metadata m ON m.book_id = b.id`

// ListLibraryBooks returns all books in a library.
func ListLibraryBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	libraryID := chi.URLParam(r, "id")

	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM libraries WHERE id = $1 AND user_id = $2)",
		libraryID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` WHERE b.library_id = $1 ORDER BY m.title`, libraryID)
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

// GetBook returns a single book by ID.
func GetBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		bookQuery+` WHERE b.id = $1 AND b.user_id = $2`, id, userID)
	b, err := scanBook(row.Scan)
	if err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	books := []models.Book{b}
	if err := attachBookRelations(r, books); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books[0])
}

func GetBookAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` WHERE b.user_id = $1`, userID)
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

// GetBookCover serves the cover image from disk.
func GetBookCover(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var coverPath, coverMime *string
	err := db.DB.QueryRow(r.Context(),
		`SELECT m.cover_path, m.cover_mime FROM book_metadata m
		JOIN books b ON b.id = m.book_id
		WHERE m.book_id = $1 AND b.user_id = $2`,
		id, userID).Scan(&coverPath, &coverMime)
	if err != nil || coverPath == nil || *coverPath == "" {
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	}

	data, err := os.ReadFile(*coverPath)
	if err != nil {
		http.Error(w, "cover file not found", http.StatusNotFound)
		return
	}

	mime := "image/jpeg"
	if coverMime != nil && *coverMime != "" {
		mime = *coverMime
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(data)
}

type metadataUpdate struct {
	Title         string   `json:"title"`
	Subtitle      string   `json:"subtitle"`
	Description   string   `json:"description"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publishedDate"`
	ISBN13        string   `json:"isbn13"`
	ISBN10        string   `json:"isbn10"`
	Language      string   `json:"language"`
	PageCount     *int     `json:"pageCount"`
	SeriesName    string   `json:"seriesName"`
	SeriesNumber  *float64 `json:"seriesNumber"`
}

type bookUpdate struct {
	Metadata   *metadataUpdate `json:"metadata"`
	Authors    *[]string       `json:"authors"`
	Categories *[]string       `json:"categories"`
	Tags       *[]string       `json:"tags"`
}

// UpdateBook updates the metadata and relations of a book.
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	row := db.DB.QueryRow(r.Context(),
		bookQuery+` WHERE b.id = $1 AND b.user_id = $2`, id, userID)
	existing, err := scanBook(row.Scan)
	if err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	var body bookUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Update metadata
	if body.Metadata != nil {
		m := body.Metadata
		_, err = db.DB.Exec(r.Context(),
			`UPDATE book_metadata SET
				title=$1, subtitle=$2, description=$3, publisher=$4, published_date=$5,
				isbn_13=$6, isbn_10=$7, language=$8, page_count=$9,
				series_name=$10, series_number=$11
			WHERE book_id = $12`,
			m.Title, nilIfEmpty(m.Subtitle),
			nilIfEmpty(m.Description), nilIfEmpty(m.Publisher), nilIfEmpty(m.PublishedDate),
			nilIfEmpty(m.ISBN13), nilIfEmpty(m.ISBN10), nilIfEmpty(m.Language), m.PageCount,
			nilIfEmpty(m.SeriesName), m.SeriesNumber,
			id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	// Update authors
	if body.Authors != nil {
		authors, err := findOrCreateAuthors(r, *body.Authors, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookAuthors(r, id, authors); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Authors = authors
	}

	// Update categories
	if body.Categories != nil {
		cats, err := findOrCreateCategories(r, *body.Categories, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookCategories(r, id, cats); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Categories = cats
	}

	// Update tags
	if body.Tags != nil {
		tagList, err := findOrCreateTags(r, *body.Tags, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookTags(r, id, tagList); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Tags = tagList
	}

	// Refetch relations if not explicitly updated
	if body.Authors == nil || body.Categories == nil || body.Tags == nil {
		books := []models.Book{existing}
		if err := attachBookRelations(r, books); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing = books[0]
	}

	// Write metadata to file if the user setting is enabled
	var writeToFile bool
	_ = db.DB.QueryRow(r.Context(),
		"SELECT write_metadata_to_file FROM user_settings WHERE user_id = $1",
		userID).Scan(&writeToFile)
	if writeToFile {
		authorNames := make([]string, len(existing.Authors))
		for i, a := range existing.Authors {
			authorNames[i] = a.Name
		}
		catNames := make([]string, len(existing.Categories))
		for i, c := range existing.Categories {
			catNames[i] = c.Name
		}
		wm := metadata.WriteMeta{
			Title:       existing.Metadata.Title,
			Authors:     authorNames,
			Description: ptrStr(existing.Metadata.Description),
			Publisher:   ptrStr(existing.Metadata.Publisher),
			Date:        ptrStr(existing.Metadata.PublishedDate),
			Language:    ptrStr(existing.Metadata.Language),
			Subject:     strings.Join(catNames, ", "),
		}
		// Best-effort: don't fail the request if file write fails
		_ = metadata.Write(existing.FilePath, wm)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DeleteBook removes a book record. Pass ?deleteFile=true to also delete the file from disk.
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	shouldDeleteFile := r.URL.Query().Get("deleteFile") == "true"

	var filePath string
	if shouldDeleteFile {
		if err := db.DB.QueryRow(r.Context(),
			"SELECT file_path FROM books WHERE id = $1 AND user_id = $2", id, userID).Scan(&filePath); err != nil {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
	}

	// Get cover path to clean up
	var coverPath *string
	_ = db.DB.QueryRow(r.Context(),
		"SELECT cover_path FROM book_metadata WHERE book_id = $1", id).Scan(&coverPath)

	result, err := db.DB.Exec(r.Context(),
		"DELETE FROM books WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() == 0 {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	if shouldDeleteFile && filePath != "" {
		os.Remove(filePath)
	}
	if coverPath != nil && *coverPath != "" {
		os.Remove(*coverPath)
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadBook serves the book file for download.
func DownloadBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	var filePath string
	if err := db.DB.QueryRow(r.Context(),
		"SELECT file_path FROM books WHERE id = $1 AND user_id = $2", id, userID).Scan(&filePath); err != nil {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)))
	http.ServeFile(w, r, filePath)
}

// ListBookShelves returns the shelf IDs a book belongs to.
func ListBookShelves(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	rows, err := db.DB.Query(r.Context(),
		"SELECT bs.shelf_id FROM book_shelves bs JOIN books b ON b.id = bs.book_id WHERE bs.book_id = $1 AND b.user_id = $2",
		id, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	shelfIDs := []string{}
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		shelfIDs = append(shelfIDs, sid)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shelfIDs)
}

// ListSeries returns distinct series names for autocomplete.
func ListSeries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	q := r.URL.Query().Get("q")

	var query string
	var args []any
	if q != "" {
		query = `SELECT DISTINCT m.series_name FROM book_metadata m
			JOIN books b ON b.id = m.book_id
			WHERE b.user_id = $1 AND m.series_name IS NOT NULL AND m.series_name ILIKE $2
			ORDER BY m.series_name`
		args = []any{userID, "%" + q + "%"}
	} else {
		query = `SELECT DISTINCT m.series_name FROM book_metadata m
			JOIN books b ON b.id = m.book_id
			WHERE b.user_id = $1 AND m.series_name IS NOT NULL
			ORDER BY m.series_name`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	series := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		series = append(series, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(series)
}

// attachBookRelations populates Authors, Categories, and Tags on a slice of books.
func attachBookRelations(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}
	if err := attachAuthors(r, books); err != nil {
		return err
	}
	if err := attachCategories(r, books); err != nil {
		return err
	}
	return attachTags(r, books)
}

// attachCategories populates the Categories field on a slice of books.
func attachCategories(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]string, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}

	rows, err := db.DB.Query(r.Context(),
		`SELECT bc.book_id, c.id, c.name, c.user_id
		FROM book_categories bc
		JOIN categories c ON c.id = bc.category_id
		WHERE bc.book_id = ANY($1)
		ORDER BY c.name`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byBook := map[string][]models.Category{}
	for rows.Next() {
		var bookID string
		var c models.Category
		if err := rows.Scan(&bookID, &c.ID, &c.Name, &c.UserID); err != nil {
			return err
		}
		byBook[bookID] = append(byBook[bookID], c)
	}

	for i := range books {
		if cats, ok := byBook[books[i].ID]; ok {
			books[i].Categories = cats
		}
	}
	return nil
}

// attachTags populates the Tags field on a slice of books.
func attachTags(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]string, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}

	rows, err := db.DB.Query(r.Context(),
		`SELECT bt.book_id, t.id, t.name, t.user_id
		FROM book_tags bt
		JOIN tags t ON t.id = bt.tag_id
		WHERE bt.book_id = ANY($1)
		ORDER BY t.name`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byBook := map[string][]models.Tag{}
	for rows.Next() {
		var bookID string
		var t models.Tag
		if err := rows.Scan(&bookID, &t.ID, &t.Name, &t.UserID); err != nil {
			return err
		}
		byBook[bookID] = append(byBook[bookID], t)
	}

	for i := range books {
		if tags, ok := byBook[books[i].ID]; ok {
			books[i].Tags = tags
		}
	}
	return nil
}

// findOrCreateCategories takes a list of category names and returns their records.
func findOrCreateCategories(r *http.Request, names []string, userID string) ([]models.Category, error) {
	cats := make([]models.Category, 0, len(names))
	for _, name := range names {
		name = trimStr(name)
		if name == "" {
			continue
		}
		var c models.Category
		err := db.DB.QueryRow(r.Context(),
			`INSERT INTO categories (name, user_id) VALUES ($1, $2)
			ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id, name, user_id`,
			name, userID).Scan(&c.ID, &c.Name, &c.UserID)
		if err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// linkBookCategories replaces all category associations for a book.
func linkBookCategories(r *http.Request, bookID string, cats []models.Category) error {
	_, err := db.DB.Exec(r.Context(), "DELETE FROM book_categories WHERE book_id = $1", bookID)
	if err != nil {
		return err
	}
	for _, c := range cats {
		_, err := db.DB.Exec(r.Context(),
			"INSERT INTO book_categories (book_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			bookID, c.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// findOrCreateTags takes a list of tag names and returns their records.
func findOrCreateTags(r *http.Request, names []string, userID string) ([]models.Tag, error) {
	tags := make([]models.Tag, 0, len(names))
	for _, name := range names {
		name = trimStr(name)
		if name == "" {
			continue
		}
		var t models.Tag
		err := db.DB.QueryRow(r.Context(),
			`INSERT INTO tags (name, user_id) VALUES ($1, $2)
			ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id, name, user_id`,
			name, userID).Scan(&t.ID, &t.Name, &t.UserID)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// linkBookTags replaces all tag associations for a book.
func linkBookTags(r *http.Request, bookID string, tags []models.Tag) error {
	_, err := db.DB.Exec(r.Context(), "DELETE FROM book_tags WHERE book_id = $1", bookID)
	if err != nil {
		return err
	}
	for _, t := range tags {
		_, err := db.DB.Exec(r.Context(),
			"INSERT INTO book_tags (book_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			bookID, t.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeCoverToDisk writes cover bytes to <libraryFolder>/.covers/<bookID>.<ext> and returns the path.
func writeCoverToDisk(libraryFolder, bookID string, coverData []byte, coverMime string) (string, error) {
	ext := ".jpg"
	switch coverMime {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}

	coverDir := filepath.Join(libraryFolder, ".covers")
	if err := os.MkdirAll(coverDir, 0755); err != nil {
		return "", err
	}

	coverPath := filepath.Join(coverDir, bookID+ext)
	if err := os.WriteFile(coverPath, coverData, 0644); err != nil {
		return "", err
	}
	return coverPath, nil
}

// MoveBooks moves books on disk to match their library's naming pattern.
// POST /api/books/move
// Body: { "bookIds": ["id1", "id2", ...] }
func MoveBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body struct {
		BookIDs []string `json:"bookIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	type moveResult struct {
		BookID string `json:"bookId"`
		Error  string `json:"error,omitempty"`
	}
	results := make([]moveResult, 0, len(body.BookIDs))

	for _, bookID := range body.BookIDs {
		res := moveResult{BookID: bookID}

		// Fetch book + metadata + library folder
		var filePath, libraryID, libraryFolder string
		var title string
		var publishedDate, publisher, language, seriesName *string
		var seriesNumber *float64
		err := db.DB.QueryRow(r.Context(),
			`SELECT b.file_path, b.library_id, l.folder,
				m.title, m.published_date, m.publisher, m.language,
				m.series_name, m.series_number
			FROM books b
			JOIN book_metadata m ON m.book_id = b.id
			JOIN libraries l ON l.id = b.library_id
			WHERE b.id = $1 AND b.user_id = $2`,
			bookID, userID).Scan(
			&filePath, &libraryID, &libraryFolder,
			&title, &publishedDate, &publisher, &language,
			&seriesName, &seriesNumber,
		)
		if err != nil {
			res.Error = "book not found"
			results = append(results, res)
			continue
		}

		// Get primary author
		var authorName *string
		_ = db.DB.QueryRow(r.Context(),
			`SELECT a.name FROM authors a
			JOIN book_authors ba ON ba.author_id = a.id
			WHERE ba.book_id = $1 LIMIT 1`, bookID).Scan(&authorName)

		ext := filepath.Ext(filePath)
		pattern := getEffectivePattern(r, libraryID, userID)
		pd := PatternData{
			Authors: "Unknown",
			Title:   title,
			Ext:     ext,
		}
		if authorName != nil && *authorName != "" {
			pd.Authors = *authorName
		}
		if publishedDate != nil && *publishedDate != "" {
			pd.Year = extractYear(*publishedDate)
		}
		if publisher != nil && *publisher != "" {
			pd.Publisher = *publisher
		}
		if language != nil && *language != "" {
			pd.Language = *language
		}
		if seriesName != nil && *seriesName != "" {
			pd.Series = *seriesName
		}
		if seriesNumber != nil {
			pd.SeriesIndex = formatSeriesIndex(*seriesNumber)
		}

		relativePath := resolveFilePattern(pattern, pd)
		destPath := filepath.Join(libraryFolder, relativePath)

		// Skip if already at the correct location
		if filepath.Clean(filePath) == filepath.Clean(destPath) {
			results = append(results, res)
			continue
		}

		destPath = uniqueDest(destPath)
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			res.Error = fmt.Sprintf("failed to create directory: %v", err)
			results = append(results, res)
			continue
		}

		if err := moveFile(filePath, destPath); err != nil {
			res.Error = fmt.Sprintf("failed to move file: %v", err)
			results = append(results, res)
			continue
		}

		// Update file_path in DB
		_, err = db.DB.Exec(r.Context(),
			"UPDATE books SET file_path = $1 WHERE id = $2",
			destPath, bookID)
		if err != nil {
			// Try to move back
			_ = moveFile(destPath, filePath)
			res.Error = "db error updating path"
			results = append(results, res)
			continue
		}

		results = append(results, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func trimStr(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
