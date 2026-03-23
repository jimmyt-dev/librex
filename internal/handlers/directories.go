package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type dirEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func ListDirectories(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir = "/"
	}

	dir = filepath.Clean(dir)
	if !filepath.IsAbs(dir) {
		http.Error(w, "path must be absolute", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "cannot read directory", http.StatusBadRequest)
		return
	}

	dirs := []dirEntry{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, dirEntry{
				Name: entry.Name(),
				Path: filepath.Join(dir, entry.Name()),
			})
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Current string     `json:"current"`
		Parent  string     `json:"parent"`
		Dirs    []dirEntry `json:"dirs"`
	}{
		Current: dir,
		Parent:  filepath.Dir(dir),
		Dirs:    dirs,
	})
}
