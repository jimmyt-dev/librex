package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StagedBook represents a file found in the bookdrop waiting to be imported.
type StagedBook struct {
	OriginalPath string `json:"originalPath"`
	FileName     string `json:"fileName"`
	Title        string `json:"title"`
	Ext          string `json:"ext"`
}

func ScanBookdrop(w http.ResponseWriter, r *http.Request) {
	// You can pass the bookdrop path via query param (e.g., ?path=/data/bookdrop)
	// Alternatively, you could fetch a user-specific bookdrop path from your DB.
	targetDir := r.URL.Query().Get("path")
	if targetDir == "" {
		// Fallback to a default bookdrop location if none is provided
		targetDir = "/Users/jimmy/Documents/Code Shit/reliquary/data/bookdrop"
	}

	cleanedDir := filepath.Clean(targetDir)

	// Validation mirroring your validateFolder logic
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

	// Define valid extensions for your media server
	validExts := map[string]bool{
		".epub": true,
		".pdf":  true,
		".mobi": true,
		".azw3": true,
		".cbz":  true,
		".cbr":  true,
	}

	var stagedBooks []StagedBook

	for _, entry := range entries {
		// Skip directories and hidden files (like .DS_Store)
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !validExts[ext] {
			continue // Skip unsupported file types
		}

		// Create a best-guess title by stripping the extension
		// Example: "Brandon Sanderson - Way of Kings.epub" -> "Brandon Sanderson - Way of Kings"
		// The frontend can parse this further if desired.
		baseName := entry.Name()
		titleGuess := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		stagedBooks = append(stagedBooks, StagedBook{
			OriginalPath: filepath.Join(cleanedDir, baseName),
			FileName:     baseName,
			Title:        titleGuess,
			Ext:          ext,
		})
	}

	// Ensure we return an empty array `[]` instead of `null` if no books are found,
	// which is much easier to map over on the frontend.
	if stagedBooks == nil {
		stagedBooks = []StagedBook{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stagedBooks)
}
