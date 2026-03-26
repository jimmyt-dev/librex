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
)

const bookCols = `b.id, b.library_id, b.user_id, b.file_path, b.added_on,
	m.title, m.subtitle, m.description, m.publisher, m.published_date,
	m.isbn_13, m.isbn_10, m.language, m.page_count,
	m.series_name, m.series_number, m.series_total, m.rating, m.cover_path, m.cover_mime`

func scanBook(scan func(dest ...any) error) (models.Book, error) {
	var b models.Book
	err := scan(
		&b.ID, &b.LibraryID, &b.UserID, &b.FilePath, &b.AddedOn,
		&b.Metadata.Title, &b.Metadata.Subtitle, &b.Metadata.Description,
		&b.Metadata.Publisher, &b.Metadata.PublishedDate,
		&b.Metadata.ISBN13, &b.Metadata.ISBN10, &b.Metadata.Language, &b.Metadata.PageCount,
		&b.Metadata.SeriesName, &b.Metadata.SeriesNumber, &b.Metadata.SeriesTotal, &b.Metadata.Rating,
		&b.Metadata.CoverPath, &b.Metadata.CoverMime,
	)
	if err == nil {
		b.Metadata.BookID = b.ID
		b.Authors = []models.Author{}
		b.Genres = []models.Genre{}
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
	SeriesTotal   *int     `json:"seriesTotal"`
	Rating        *int     `json:"rating"`
}

type bookUpdate struct {
	Metadata *metadataUpdate `json:"metadata"`
	Authors  *[]string       `json:"authors"`
	Genres   *[]string       `json:"genres"`
	Tags     *[]string       `json:"tags"`
}

// UpdateBook updates the metadata and relations of a book.
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	row := tx.QueryRow(r.Context(),
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
		_, err = tx.Exec(r.Context(),
			`UPDATE book_metadata SET
				title=$1, subtitle=$2, description=$3, publisher=$4, published_date=$5,
				isbn_13=$6, isbn_10=$7, language=$8, page_count=$9,
				series_name=$10, series_number=$11, series_total=$12, rating=$13
			WHERE book_id = $14`,
			m.Title, nilIfEmpty(m.Subtitle),
			nilIfEmpty(m.Description), nilIfEmpty(m.Publisher), nilIfEmpty(m.PublishedDate),
			nilIfEmpty(m.ISBN13), nilIfEmpty(m.ISBN10), nilIfEmpty(m.Language), m.PageCount,
			nilIfEmpty(m.SeriesName), m.SeriesNumber, m.SeriesTotal, m.Rating,
			id)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		// Reflect new values for response
		existing.Metadata.Title = m.Title
		existing.Metadata.Subtitle = nilIfEmpty(m.Subtitle)
		existing.Metadata.Description = nilIfEmpty(m.Description)
		existing.Metadata.Publisher = nilIfEmpty(m.Publisher)
		existing.Metadata.PublishedDate = nilIfEmpty(m.PublishedDate)
		existing.Metadata.ISBN13 = nilIfEmpty(m.ISBN13)
		existing.Metadata.ISBN10 = nilIfEmpty(m.ISBN10)
		existing.Metadata.Language = nilIfEmpty(m.Language)
		existing.Metadata.PageCount = m.PageCount
		existing.Metadata.SeriesName = nilIfEmpty(m.SeriesName)
		existing.Metadata.SeriesNumber = m.SeriesNumber
		existing.Metadata.SeriesTotal = m.SeriesTotal
		existing.Metadata.Rating = m.Rating
	}

	// Update authors
	if body.Authors != nil {
		authors, err := findOrCreateAuthorsTX(r, tx, *body.Authors, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookAuthors(r, tx, id, authors); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Authors = authors
	}

	// Update genres
	if body.Genres != nil {
		genres, err := findOrCreateGenresTX(r, tx, *body.Genres, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookGenres(r, tx, id, genres); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Genres = genres
	}

	// Update tags
	if body.Tags != nil {
		tagList, err := findOrCreateTagsTX(r, tx, *body.Tags, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if err := linkBookTags(r, tx, id, tagList); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		existing.Tags = tagList
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Refetch relations if not explicitly updated to ensure consistency
	if body.Authors == nil || body.Genres == nil || body.Tags == nil {
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
		genreNames := make([]string, len(existing.Genres))
		for i, g := range existing.Genres {
			genreNames[i] = g.Name
		}
		wm := metadata.WriteMeta{
			Title:       existing.Metadata.Title,
			Authors:     &authorNames,
			Description: existing.Metadata.Description,
			Publisher:   existing.Metadata.Publisher,
			Date:        existing.Metadata.PublishedDate,
			Language:    existing.Metadata.Language,
			Subjects:    &genreNames,
		}
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

	// Cleanup orphans after deletion
	_ = CleanupOrphanAuthors(r, db.DB, userID)
	_ = CleanupOrphanGenres(r, db.DB, userID)
	_ = CleanupOrphanTags(r, db.DB, userID)

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

// ListPublishers returns a list of publishers for the user's books.
func ListPublishers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	q := r.URL.Query().Get("q")

	var query string
	var args []any
	if q != "" {
		query = `SELECT DISTINCT m.publisher FROM book_metadata m
			JOIN books b ON b.id = m.book_id
			WHERE b.user_id = $1 AND m.publisher IS NOT NULL AND m.publisher ILIKE $2
			ORDER BY m.publisher`
		args = []any{userID, "%" + q + "%"}
	} else {
		query = `SELECT DISTINCT m.publisher FROM book_metadata m
			JOIN books b ON b.id = m.book_id
			WHERE b.user_id = $1 AND m.publisher IS NOT NULL
			ORDER BY m.publisher`
		args = []any{userID}
	}

	rows, err := db.DB.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	publishers := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		publishers = append(publishers, name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publishers)
}

// attachBookRelations populates Authors, Genres, and Tags on a slice of books.
func attachBookRelations(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}
	if err := attachAuthors(r, books); err != nil {
		return err
	}
	if err := attachGenres(r, books); err != nil {
		return err
	}
	return attachTags(r, books)
}

// attachGenres populates the Genres field on a slice of books.
func attachGenres(r *http.Request, books []models.Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]string, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}

	rows, err := db.DB.Query(r.Context(),
		`SELECT bg.book_id, g.id, g.name, g.user_id
		FROM book_genres bg
		JOIN genres g ON g.id = bg.genre_id
		WHERE bg.book_id = ANY($1)
		ORDER BY g.name`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byBook := map[string][]models.Genre{}
	for rows.Next() {
		var bookID string
		var g models.Genre
		if err := rows.Scan(&bookID, &g.ID, &g.Name, &g.UserID); err != nil {
			return err
		}
		byBook[bookID] = append(byBook[bookID], g)
	}

	for i := range books {
		if genres, ok := byBook[books[i].ID]; ok {
			books[i].Genres = genres
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

// findOrCreateGenres takes a list of genre names and returns their records.
func findOrCreateGenres(r *http.Request, names []string, userID string) ([]models.Genre, error) {
	return findOrCreateGenresTX(r, db.DB, names, userID)
}

func findOrCreateGenresTX(r *http.Request, q db.DBTX, names []string, userID string) ([]models.Genre, error) {
	result := make([]models.Genre, 0, len(names))
	for _, name := range names {
		name = trimStr(name)
		if name == "" {
			continue
		}
		var g models.Genre
		err := q.QueryRow(r.Context(),
			`INSERT INTO genres (name, user_id) VALUES ($1, $2)
			ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id, name, user_id`,
			name, userID).Scan(&g.ID, &g.Name, &g.UserID)
		if err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, nil
}

// linkBookGenres replaces all genre associations for a book.
func linkBookGenres(r *http.Request, q db.DBTX, bookID string, genres []models.Genre) error {
	_, err := q.Exec(r.Context(), "DELETE FROM book_genres WHERE book_id = $1", bookID)
	if err != nil {
		return err
	}
	for _, g := range genres {
		_, err := q.Exec(r.Context(),
			"INSERT INTO book_genres (book_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			bookID, g.ID)
		if err != nil {
			return err
		}
	}
	return CleanupOrphanGenres(r, q, middleware.GetUserID(r))
}

func CleanupOrphanGenres(r *http.Request, q db.DBTX, userID string) error {
	_, err := q.Exec(r.Context(),
		`DELETE FROM genres
		 WHERE user_id = $1 AND id NOT IN (SELECT genre_id FROM book_genres)`,
		userID)
	return err
}

// linkBookTags replaces all tag associations for a book.
func linkBookTags(r *http.Request, q db.DBTX, bookID string, tags []models.Tag) error {
	_, err := q.Exec(r.Context(), "DELETE FROM book_tags WHERE book_id = $1", bookID)
	if err != nil {
		return err
	}
	for _, t := range tags {
		_, err := q.Exec(r.Context(),
			"INSERT INTO book_tags (book_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			bookID, t.ID)
		if err != nil {
			return err
		}
	}
	return CleanupOrphanTags(r, q, middleware.GetUserID(r))
}

func CleanupOrphanTags(r *http.Request, q db.DBTX, userID string) error {
	_, err := q.Exec(r.Context(),
		`DELETE FROM tags
		 WHERE user_id = $1 AND id NOT IN (SELECT tag_id FROM book_tags)`,
		userID)
	return err
}

// findOrCreateTags takes a list of tag names and returns their records.
func findOrCreateTags(r *http.Request, names []string, userID string) ([]models.Tag, error) {
	return findOrCreateTagsTX(r, db.DB, names, userID)
}

func findOrCreateTagsTX(r *http.Request, q db.DBTX, names []string, userID string) ([]models.Tag, error) {
	tags := make([]models.Tag, 0, len(names))
	for _, name := range names {
		name = trimStr(name)
		if name == "" {
			continue
		}
		var t models.Tag
		err := q.QueryRow(r.Context(),
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

		// COPY first, then update DB, then delete old file. This is more robust than Rename.
		if err := copyFile(filePath, destPath); err != nil {
			res.Error = fmt.Sprintf("failed to copy file: %v", err)
			results = append(results, res)
			continue
		}

		// Update file_path in DB
		_, err = db.DB.Exec(r.Context(),
			"UPDATE books SET file_path = $1 WHERE id = $2 AND user_id = $3",
			destPath, bookID, userID)
		if err != nil {
			// Clean up the new copy on failure
			os.Remove(destPath)
			res.Error = "db error updating path"
			results = append(results, res)
			continue
		}

		// Success! Now safe to remove the original file.
		os.Remove(filePath)
		results = append(results, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type bulkUpdateBody struct {
	BookIDs []string `json:"bookIds"`
	// Metadata fields — nil means "don't change"
	SeriesName  *string `json:"seriesName"`
	Publisher   *string `json:"publisher"`
	Language    *string `json:"language"`
	SeriesTotal *int    `json:"seriesTotal"`
	Rating      *int    `json:"rating"`
	// Array fields with mode
	Authors     *[]string `json:"authors"`
	AuthorsMode string    `json:"authorsMode"` // "replace" or "merge"
	Genres      *[]string `json:"genres"`
	GenresMode  string    `json:"genresMode"`
	Tags        *[]string `json:"tags"`
	TagsMode    string    `json:"tagsMode"`
}

// BulkUpdateBooks updates metadata across multiple books at once.
// POST /api/books/bulk-update
func BulkUpdateBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body bulkUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.BookIDs) == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Pre-resolve array values once
	var newAuthors []models.Author
	var newGenres []models.Genre
	var newTags []models.Tag
	var err error

	if body.Authors != nil {
		newAuthors, err = findOrCreateAuthors(r, *body.Authors, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}
	if body.Genres != nil {
		newGenres, err = findOrCreateGenres(r, *body.Genres, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}
	if body.Tags != nil {
		newTags, err = findOrCreateTags(r, *body.Tags, userID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	tx, err := db.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	type result struct {
		BookID string `json:"bookId"`
		Error  string `json:"error,omitempty"`
	}
	results := make([]result, 0, len(body.BookIDs))

	for i, bookID := range body.BookIDs {
		res := result{BookID: bookID}

		// Use savepoints to prevent a single failure from poisoning the entire transaction
		spName := fmt.Sprintf("sp_%d", i)
		_, _ = tx.Exec(r.Context(), "SAVEPOINT "+spName)

		// Verify ownership
		var exists bool
		err := tx.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM books WHERE id = $1 AND user_id = $2)",
			bookID, userID).Scan(&exists)
		if err != nil || !exists {
			_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
			res.Error = "book not found"
			results = append(results, res)
			continue
		}

		// Update scalar metadata fields
		if body.SeriesName != nil || body.Publisher != nil || body.Language != nil || body.SeriesTotal != nil || body.Rating != nil {
			_, err := tx.Exec(r.Context(),
				`UPDATE book_metadata SET
					series_name  = COALESCE($1, series_name),
					publisher    = COALESCE($2, publisher),
					language     = COALESCE($3, language),
					series_total = COALESCE($4, series_total),
					rating       = COALESCE($5, rating)
				WHERE book_id = $6`,
				nilIfEmptyPtr(body.SeriesName), nilIfEmptyPtr(body.Publisher), nilIfEmptyPtr(body.Language),
				body.SeriesTotal, body.Rating,
				bookID)
			if err != nil {
				_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
				res.Error = "db error updating metadata"
				results = append(results, res)
				continue
			}
		}

		// Authors
		if body.Authors != nil {
			var finalAuthors []models.Author
			if body.AuthorsMode == "merge" {
				existing, err := getBookAuthors(r, tx, bookID)
				if err != nil {
					_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
					res.Error = "db error fetching authors"
					results = append(results, res)
					continue
				}
				finalAuthors = mergeAuthors(existing, newAuthors)
			} else {
				finalAuthors = newAuthors
			}
			if err := linkBookAuthors(r, tx, bookID, finalAuthors); err != nil {
				_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
				res.Error = "db error linking authors"
				results = append(results, res)
				continue
			}
		}

		// Genres
		if body.Genres != nil {
			var finalGenres []models.Genre
			if body.GenresMode == "merge" {
				existing, err := getBookGenres(r, tx, bookID)
				if err != nil {
					_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
					res.Error = "db error fetching genres"
					results = append(results, res)
					continue
				}
				finalGenres = mergeGenres(existing, newGenres)
			} else {
				finalGenres = newGenres
			}
			if err := linkBookGenres(r, tx, bookID, finalGenres); err != nil {
				_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
				res.Error = "db error linking genres"
				results = append(results, res)
				continue
			}
		}

		// Tags
		if body.Tags != nil {
			var finalTags []models.Tag
			if body.TagsMode == "merge" {
				existing, err := getBookTags(r, tx, bookID)
				if err != nil {
					_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
					res.Error = "db error fetching tags"
					results = append(results, res)
					continue
				}
				finalTags = mergeTags(existing, newTags)
			} else {
				finalTags = newTags
			}
			if err := linkBookTags(r, tx, bookID, finalTags); err != nil {
				_, _ = tx.Exec(r.Context(), "ROLLBACK TO SAVEPOINT "+spName)
				res.Error = "db error linking tags"
				results = append(results, res)
				continue
			}
		}

		_, _ = tx.Exec(r.Context(), "RELEASE SAVEPOINT "+spName)
		results = append(results, res)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Write metadata to file for successfully updated books, if setting is enabled
	var writeToFile bool
	_ = db.DB.QueryRow(r.Context(),
		"SELECT write_metadata_to_file FROM user_settings WHERE user_id = $1",
		userID).Scan(&writeToFile)
	if writeToFile {
		for _, res := range results {
			if res.Error != "" {
				continue
			}
			row := db.DB.QueryRow(r.Context(), bookQuery+` WHERE b.id = $1 AND b.user_id = $2`, res.BookID, userID)
			book, err := scanBook(row.Scan)
			if err != nil {
				continue
			}
			// Only write the fields that were actually in the bulk-edit request.
			// Leaving a WriteMeta field nil means "don't touch that element in the file."
			wm := metadata.WriteMeta{}
			if body.Publisher != nil {
				wm.Publisher = book.Metadata.Publisher
			}
			if body.Language != nil {
				wm.Language = book.Metadata.Language
			}
			if body.Authors != nil {
				names := make([]string, len(book.Authors))
				for i, a := range book.Authors {
					names[i] = a.Name
				}
				wm.Authors = &names
			}
			if body.Genres != nil {
				genres := make([]string, len(book.Genres))
				for i, g := range book.Genres {
					genres[i] = g.Name
				}
				wm.Subjects = &genres
			}
			_ = metadata.Write(book.FilePath, wm)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func nilIfEmptyPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func getBookAuthors(r *http.Request, q db.DBTX, bookID string) ([]models.Author, error) {
	rows, err := q.Query(r.Context(),
		`SELECT a.id, a.name, a.user_id FROM authors a JOIN book_authors ba ON ba.author_id = a.id WHERE ba.book_id = $1`,
		bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Author
	for rows.Next() {
		var a models.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.UserID); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func getBookGenres(r *http.Request, q db.DBTX, bookID string) ([]models.Genre, error) {
	rows, err := q.Query(r.Context(),
		`SELECT g.id, g.name, g.user_id FROM genres g JOIN book_genres bg ON bg.genre_id = g.id WHERE bg.book_id = $1`,
		bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Genre
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.UserID); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, nil
}

func getBookTags(r *http.Request, q db.DBTX, bookID string) ([]models.Tag, error) {
	rows, err := q.Query(r.Context(),
		`SELECT t.id, t.name, t.user_id FROM tags t JOIN book_tags bt ON bt.tag_id = t.id WHERE bt.book_id = $1`,
		bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.UserID); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func mergeAuthors(existing, incoming []models.Author) []models.Author {
	seen := map[string]bool{}
	result := make([]models.Author, 0, len(existing)+len(incoming))
	for _, a := range existing {
		seen[a.ID] = true
		result = append(result, a)
	}
	for _, a := range incoming {
		if !seen[a.ID] {
			result = append(result, a)
		}
	}
	return result
}

func mergeGenres(existing, incoming []models.Genre) []models.Genre {
	seen := map[string]bool{}
	result := make([]models.Genre, 0, len(existing)+len(incoming))
	for _, g := range existing {
		seen[g.ID] = true
		result = append(result, g)
	}
	for _, g := range incoming {
		if !seen[g.ID] {
			result = append(result, g)
		}
	}
	return result
}

func mergeTags(existing, incoming []models.Tag) []models.Tag {
	seen := map[string]bool{}
	result := make([]models.Tag, 0, len(existing)+len(incoming))
	for _, t := range existing {
		seen[t.ID] = true
		result = append(result, t)
	}
	for _, t := range incoming {
		if !seen[t.ID] {
			result = append(result, t)
		}
	}
	return result
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
