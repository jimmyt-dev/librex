package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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
	feed := OPDSFeed{
		ID:      baseURL + "/opds",
		Title:   "Reliquary OPDS Catalog",
		Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
		Entries: []OPDSEntry{
			{
				ID:      "all-books",
				Title:   "All Books",
				Updated: time.Now().Format(time.RFC3339),
				Links: []OPDSLink{
					{Rel: "subsection", Href: baseURL + "/opds/all.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
			{
				ID:      "new-books",
				Title:   "Newest Books",
				Updated: time.Now().Format(time.RFC3339),
				Links: []OPDSLink{
					{Rel: "subsection", Href: baseURL + "/opds/new.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` WHERE b.user_id = $1 ORDER BY m.title`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			continue
		}
		books = append(books, b)
	}

	if err := attachBookRelations(r, books); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	feed := OPDSFeed{
		ID:      baseURL + "/opds/all.xml",
		Title:   "All Books",
		Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/all.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, b := range books {
		entry := OPDSEntry{
			ID:      fmt.Sprintf("book:%s", b.ID),
			Title:   b.Metadata.Title,
			Updated: b.AddedOn.Format(time.RFC3339),
			Summary: *b.Metadata.Description,
		}

		for _, a := range b.Authors {
			entry.Authors = append(entry.Authors, OPDSAuthor{Name: a.Name})
		}

		// Download link
		entry.Links = append(entry.Links, OPDSLink{
			Rel:  "http://opds-spec.org/acquisition",
			Href: fmt.Sprintf("%s/api/books/%s/download", baseURL, b.ID),
			Type: getMimeType(b.FilePath),
		})

		// Cover link
		if b.Metadata.CoverPath != nil {
			entry.Links = append(entry.Links, OPDSLink{
				Rel:  "http://opds-spec.org/image",
				Href: fmt.Sprintf("%s/api/books/%s/cover", baseURL, b.ID),
				Type: *b.Metadata.CoverMime,
			})
			entry.Links = append(entry.Links, OPDSLink{
				Rel:  "http://opds-spec.org/image/thumbnail",
				Href: fmt.Sprintf("%s/api/books/%s/cover", baseURL, b.ID),
				Type: *b.Metadata.CoverMime,
			})
		}

		feed.Entries = append(feed.Entries, entry)
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func GetOPDSNew(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	baseURL := getBaseURL(r)

	rows, err := db.DB.Query(r.Context(),
		bookQuery+` WHERE b.user_id = $1 ORDER BY b.added_on DESC LIMIT 50`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		b, err := scanBook(rows.Scan)
		if err != nil {
			continue
		}
		books = append(books, b)
	}

	if err := attachBookRelations(r, books); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	feed := OPDSFeed{
		ID:      baseURL + "/opds/new.xml",
		Title:   "Newest Books",
		Updated: time.Now().Format(time.RFC3339),
		Links: []OPDSLink{
			{Rel: "self", Href: baseURL + "/opds/new.xml", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "up", Href: baseURL + "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, b := range books {
		entry := OPDSEntry{
			ID:      fmt.Sprintf("book:%s", b.ID),
			Title:   b.Metadata.Title,
			Updated: b.AddedOn.Format(time.RFC3339),
			Summary: *b.Metadata.Description,
		}

		for _, a := range b.Authors {
			entry.Authors = append(entry.Authors, OPDSAuthor{Name: a.Name})
		}

		entry.Links = append(entry.Links, OPDSLink{
			Rel:  "http://opds-spec.org/acquisition",
			Href: fmt.Sprintf("%s/api/books/%s/download", baseURL, b.ID),
			Type: getMimeType(b.FilePath),
		})

		if b.Metadata.CoverPath != nil {
			entry.Links = append(entry.Links, OPDSLink{
				Rel:  "http://opds-spec.org/image",
				Href: fmt.Sprintf("%s/api/books/%s/cover", baseURL, b.ID),
				Type: *b.Metadata.CoverMime,
			})
		}

		feed.Entries = append(feed.Entries, entry)
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(feed)
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
		scheme = xfp
	}

	host := r.Host
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		host = xfh
	}

	return fmt.Sprintf("%s://%s", scheme, host)
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
