package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AllowedRoots returns the list of directories that the application is allowed to access.
// In a real app, this should be configurable via env or database.
func AllowedRoots() []string {
	roots := os.Getenv("RELIQUARY_ALLOWED_ROOTS")
	if roots == "" {
		// If not set, we default to the entire filesystem for now,
		// but this function provides the hook to restrict it.
		return []string{"/"}
	}
	return strings.Split(roots, ",")
}

// ValidatePath ensures that a given path is within the allowed root directories.
func ValidatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path must be absolute")
	}

	allowed := false
	for _, root := range AllowedRoots() {
		rel, err := filepath.Rel(root, cleaned)
		if err == nil && !strings.HasPrefix(rel, "..") {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("path is outside of allowed root directories")
	}

	return cleaned, nil
}

type handlerError struct {
	msg  string
	code int
}

func copyFile(src, dst string) error {
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
	return out.Close()
}
