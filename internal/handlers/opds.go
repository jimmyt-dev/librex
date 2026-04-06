package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"librex/internal/db"
	"librex/internal/middleware"
	"librex/internal/models"
)

type OPDSFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Icon    string      `xml:"icon,omitempty"`
	Links   []OPDSLink  `xml:"link"`
	Entries []OPDSEntry `xml:"entry"`
}

type OPDSLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type OPDSEntry struct {
	ID      string       `xml:"id"`
	Title   string       `xml:"title"`
	Updated string       `xml:"updated"`
	Authors []OPDSAuthor `xml:"author"`
	Summary string       `xml:"summary,omitempty"`
	Links   []OPDSLink   `xml:"link"`
}

type OPDSAuthor struct {
	Name string `xml:"name"`
}

func GetOPDSRoot(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL(r)
	now := time.Now().Format(time.RFC3339)
	feed := OPDSFeed{
		ID:      baseURL + "/opds",
		Title:   "Librex OPDS Catalog",
		Updated: now,
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
		Entries: []OPDSEntry{
			{
				ID: "all-books", Title: "All Books", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/all.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
			},
			{
				ID: "new-books", Title: "Recently Added", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/new.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
			},
			{
				ID: "libraries", Title: "Libraries", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/libraries.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
			{
				ID: "shelves", Title: "Shelves", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/shelves.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
			{
				ID: "authors", Title: "Authors", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/authors.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
			{
				ID: "series", Title: "Series", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/series.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
			{
				ID: "random", Title: "Surprise Me (Random)", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/random.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
			},
			{
				ID: "magic", Title: "Magic Shelves (Coming Soon)", Updated: now,
				Links: []OPDSLink{{Rel: "subsection", Href: baseURL + "/opds/magic.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
		},
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	renderBookList(w, r, "All Books", "/opds/all.xml",
		bookQuery+` WHERE b.user_id = $1 ORDER BY m.title`, userID)
}

func GetOPDSNew(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	renderBookList(w, r, "Recently Added", "/opds/new.xml",
		bookQuery+` WHERE b.user_id = $1 ORDER BY b.added_on DESC LIMIT 50`, userID)
}

func GetOPDSRandom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	renderBookList(w, r, "Surprise Me", "/opds/random.xml",
		bookQuery+` WHERE b.user_id = $1 ORDER BY RANDOM() LIMIT 25`, userID)
}

func GetOPDSLibrariies(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)
	rows, err := db.DB.Query(r.Context(), "SELECT id, name FROM libraries WHERE user_id = $1 ORDER BY name", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	feed := OPDSFeed{
		ID: baseURL + "/opds/libraries.xml", Title: "Libraries", Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/libraries.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		feed.Entries = append(feed.Entries, OPDSEntry{
			ID: "lib:" + id, Title: name, Updated: feed.Updated,
			Links: []OPDSLink{{Rel: "subsection", Href: fmt.Sprintf("%s/opds/libraries/%s.xml", baseURL, id), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		})
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSLibraryBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	renderBookList(w, r, "Library", "/opds/libraries/"+id+".xml",
		bookQuery+` WHERE b.library_id = $1 AND b.user_id = $2 ORDER BY m.title`, id, userID)
}

func GetOPDSShelves(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)
	rows, err := db.DB.Query(r.Context(), "SELECT id, name FROM shelves WHERE user_id = $1 ORDER BY name", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	feed := OPDSFeed{
		ID: baseURL + "/opds/shelves.xml", Title: "Shelves", Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/shelves.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		feed.Entries = append(feed.Entries, OPDSEntry{
			ID: "shelf:" + id, Title: name, Updated: feed.Updated,
			Links: []OPDSLink{{Rel: "subsection", Href: fmt.Sprintf("%s/opds/shelves/%s.xml", baseURL, id), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		})
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSShelfBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	renderBookList(w, r, "Shelf", "/opds/shelves/"+id+".xml",
		bookQuery+` JOIN book_shelves bs ON bs.book_id = b.id WHERE bs.shelf_id = $1 AND b.user_id = $2 ORDER BY m.title`, id, userID)
}

func GetOPDSAuthors(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)
	rows, err := db.DB.Query(r.Context(), "SELECT id, name FROM authors WHERE user_id = $1 ORDER BY name", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	feed := OPDSFeed{
		ID: baseURL + "/opds/authors.xml", Title: "Authors", Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/authors.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		feed.Entries = append(feed.Entries, OPDSEntry{
			ID: "author:" + id, Title: name, Updated: feed.Updated,
			Links: []OPDSLink{{Rel: "subsection", Href: fmt.Sprintf("%s/opds/authors/%s.xml", baseURL, id), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		})
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSAuthorBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id := chi.URLParam(r, "id")
	renderBookList(w, r, "Author", "/opds/authors/"+id+".xml",
		bookQuery+` JOIN book_authors ba ON ba.book_id = b.id WHERE ba.author_id = $1 AND b.user_id = $2 ORDER BY m.title`, id, userID)
}

func GetOPDSSeries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)
	rows, err := db.DB.Query(r.Context(), "SELECT DISTINCT series_name FROM book_metadata m JOIN books b ON b.id = m.book_id WHERE b.user_id = $1 AND series_name IS NOT NULL AND series_name != '' ORDER BY series_name", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	feed := OPDSFeed{
		ID: baseURL + "/opds/series.xml", Title: "Series", Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/series.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		feed.Entries = append(feed.Entries, OPDSEntry{
			ID: "series:" + name, Title: name, Updated: feed.Updated,
			Links: []OPDSLink{{Rel: "subsection", Href: fmt.Sprintf("%s/opds/series/%s.xml", baseURL, name), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		})
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSSeriesBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	name := chi.URLParam(r, "name")
	renderBookList(w, r, "Series: "+name, "/opds/series/"+name+".xml",
		bookQuery+` WHERE m.series_name = $1 AND b.user_id = $2 ORDER BY m.series_number ASC, m.title`, name, userID)
}

func GetOPDSMagic(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL(r)
	feed := OPDSFeed{
		ID: baseURL + "/opds/magic.xml", Title: "Magic Shelves", Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/magic.xml", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
		Entries: []OPDSEntry{
			{ID: "magic:favorites", Title: "Favorites (Coming Soon)", Updated: time.Now().Format(time.RFC3339)},
			{ID: "magic:mustread", Title: "Must Read (Coming Soon)", Updated: time.Now().Format(time.RFC3339)},
		},
	}
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

// Internal Helper to render book entries from a query
func renderBookList(w http.ResponseWriter, r *http.Request, title, path string, query string, args ...any) {
	baseURL := getBaseURL(r)
	rows, err := db.DB.Query(r.Context(), query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err == nil {
			books = append(books, b)
		}
	}
	attachBookRelations(r, books)

	feed := OPDSFeed{
		ID: baseURL + path, Title: title, Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + path, Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, b := range books {
		entry := OPDSEntry{
			ID: fmt.Sprintf("book:%s", b.ID), Title: b.Metadata.Title,
			Updated: b.AddedOn.Format(time.RFC3339), Summary: getString(b.Metadata.Description),
		}
		for _, a := range b.Authors {
			entry.Authors = append(entry.Authors, OPDSAuthor{Name: a.Name})
		}
		entry.Links = append(entry.Links, OPDSLink{
			Rel: "http://opds-spec.org/acquisition", Href: fmt.Sprintf("%s/opds/books/%s/download", baseURL, b.ID), Type: getMimeType(b.FilePath),
		})
		if b.Metadata.CoverPath != nil {
			mime := getString(b.Metadata.CoverMime)
			if mime == "" {
				mime = "image/jpeg"
			}
			entry.Links = append(entry.Links, OPDSLink{Rel: "http://opds-spec.org/image", Href: fmt.Sprintf("%s/opds/books/%s/cover", baseURL, b.ID), Type: mime})
			entry.Links = append(entry.Links, OPDSLink{Rel: "http://opds-spec.org/image/thumbnail", Href: fmt.Sprintf("%s/opds/books/%s/cover", baseURL, b.ID), Type: mime})
		}
		feed.Entries = append(feed.Entries, entry)
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func getBaseURL(r *http.Request) string {
	if origin := os.Getenv("ORIGIN"); origin != "" {
		return strings.TrimSuffix(origin, "/")
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		host = xfh
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".epub":
		return "application/epub+zip"
	case ".pdf":
		return "application/pdf"
	case ".mobi":
		return "application/x-mobipocket-ebook"
	case ".azw3":
		return "application/x-mobi8-ebook"
	case ".cbz":
		return "application/x-cbz"
	default:
		return "application/octet-stream"
	}
}
