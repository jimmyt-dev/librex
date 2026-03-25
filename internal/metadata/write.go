package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// WriteMeta holds the fields to write into a book file's metadata.
type WriteMeta struct {
	Title       string
	Authors     []string // dc:creator entries
	Description string
	Publisher   string
	Date        string // publication date
	Language    string
	Subject     string // comma-separated categories/subjects
}

// Write updates metadata inside a book file (EPUB or PDF).
// Returns nil if the format is unsupported (no-op).
func Write(filePath string, meta WriteMeta) error {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".epub":
		return writeEPUB(filePath, meta)
	default:
		// PDF metadata writing is complex and risky; skip for now
		return nil
	}
}

// writeEPUB modifies the OPF metadata inside an EPUB (ZIP) file in-place.
func writeEPUB(filePath string, meta WriteMeta) error {
	// Read the entire zip into memory
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read epub: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open epub zip: %w", err)
	}

	// Find OPF path from container.xml
	opfPath := ""
	for _, f := range r.File {
		if strings.EqualFold(f.Name, "META-INF/container.xml") {
			rc, err := f.Open()
			if err != nil {
				break
			}
			var c container
			if err := xml.NewDecoder(rc).Decode(&c); err == nil && len(c.Rootfile) > 0 {
				opfPath = c.Rootfile[0].FullPath
			}
			rc.Close()
			break
		}
	}
	if opfPath == "" {
		return fmt.Errorf("no OPF file found in EPUB")
	}

	// Read and parse the OPF
	var opfData []byte
	for _, f := range r.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("open opf: %w", err)
			}
			opfData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("read opf: %w", err)
			}
			break
		}
	}
	if opfData == nil {
		return fmt.Errorf("OPF file not found at %s", opfPath)
	}

	// Modify the OPF XML
	newOPF, err := patchOPFMetadata(opfData, meta)
	if err != nil {
		return fmt.Errorf("patch opf: %w", err)
	}

	// Rewrite the ZIP with the modified OPF
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for _, f := range r.File {
		header := f.FileHeader
		writer, err := w.CreateHeader(&header)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", f.Name, err)
		}

		if f.Name == opfPath {
			// Write modified OPF
			if _, err := writer.Write(newOPF); err != nil {
				return fmt.Errorf("write modified opf: %w", err)
			}
		} else {
			// Copy original file
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("open zip entry %s: %w", f.Name, err)
			}
			if _, err := io.Copy(writer, rc); err != nil {
				rc.Close()
				return fmt.Errorf("copy zip entry %s: %w", f.Name, err)
			}
			rc.Close()
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}

	// Write back atomically: write to temp, then rename
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// patchOPFMetadata does a targeted find-and-replace on the OPF XML to update
// Dublin Core metadata fields without disturbing the rest of the document.
func patchOPFMetadata(opfData []byte, meta WriteMeta) ([]byte, error) {
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, err
	}

	m := &pkg.Metadata

	if meta.Title != "" {
		m.Title = []string{meta.Title}
	}
	if len(meta.Authors) > 0 {
		m.Creator = meta.Authors
	}
	if meta.Description != "" {
		m.Description = []string{meta.Description}
	} else {
		m.Description = nil
	}
	if meta.Publisher != "" {
		m.Publisher = []string{meta.Publisher}
	} else {
		m.Publisher = nil
	}
	if meta.Date != "" {
		m.Date = []string{meta.Date}
	} else {
		m.Date = nil
	}
	if meta.Language != "" {
		m.Language = []string{meta.Language}
	}
	if meta.Subject != "" {
		subjects := strings.Split(meta.Subject, ",")
		var trimmed []string
		for _, s := range subjects {
			s = strings.TrimSpace(s)
			if s != "" {
				trimmed = append(trimmed, s)
			}
		}
		m.Subject = trimmed
	} else {
		m.Subject = nil
	}

	out, err := xml.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, err
	}

	// Prepend XML declaration
	return append([]byte(xml.Header), out...), nil
}
