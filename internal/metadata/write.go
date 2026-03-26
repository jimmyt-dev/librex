package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

// WriteMeta holds the fields to write into a book file's metadata.
// Pointer fields: nil = don't touch that element in the file.
// Non-pointer strings: "" = don't touch.
type WriteMeta struct {
	Title       string    // "" = don't touch
	Authors     *[]string // nil = don't touch; &[]string{} = remove all creators
	Description *string   // nil = don't touch; ptr("") = remove
	Publisher   *string
	Date        *string // publication date
	Language    *string
	Subjects    *[]string // nil = don't touch; &[]string{} = remove all subjects
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

	// Read the OPF
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

	// Patch only the metadata section — never re-serialize the whole OPF
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
			if _, err := writer.Write(newOPF); err != nil {
				return fmt.Errorf("write modified opf: %w", err)
			}
		} else {
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

	// Write back atomically
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

// metaOpenRe matches the opening <metadata ...> tag (with optional namespace prefix).
var metaOpenRe = regexp.MustCompile(`<(?:[a-zA-Z0-9_-]+:)?metadata(?:\s[^>]*)?>`)

// metaCloseRe matches the closing </metadata> tag.
var metaCloseRe = regexp.MustCompile(`</(?:[a-zA-Z0-9_-]+:)?metadata>`)

// patchOPFMetadata finds the <metadata>…</metadata> block in the OPF and
// surgically replaces only the DC elements we care about.  Everything else in
// the file (spine, manifest, guide, package attributes, EPUB3 <meta> elements,
// etc.) is preserved byte-for-byte.
func patchOPFMetadata(opfData []byte, meta WriteMeta) ([]byte, error) {
	s := string(opfData)

	openLoc := metaOpenRe.FindStringIndex(s)
	if openLoc == nil {
		return nil, fmt.Errorf("no <metadata> element found in OPF")
	}
	closeLoc := metaCloseRe.FindStringIndex(s[openLoc[1]:])
	if closeLoc == nil {
		return nil, fmt.Errorf("no </metadata> closing tag found in OPF")
	}

	// Boundaries
	openTag := s[openLoc[0]:openLoc[1]]
	inner := s[openLoc[1] : openLoc[1]+closeLoc[0]]
	closeTag := s[openLoc[1]+closeLoc[0] : openLoc[1]+closeLoc[1]]
	before := s[:openLoc[0]]
	after := s[openLoc[1]+closeLoc[1]:]

	// Patch individual DC fields — nil pointer = leave untouched
	if meta.Title != "" {
		inner = replaceDCField(inner, "title", []string{meta.Title})
	}
	if meta.Authors != nil {
		inner = replaceDCField(inner, "creator", *meta.Authors)
	}
	if meta.Description != nil {
		inner = replaceDCField(inner, "description", strSliceIfNotEmpty(*meta.Description))
	}
	if meta.Publisher != nil {
		inner = replaceDCField(inner, "publisher", strSliceIfNotEmpty(*meta.Publisher))
	}
	if meta.Date != nil {
		inner = replaceDCField(inner, "date", strSliceIfNotEmpty(*meta.Date))
	}
	if meta.Language != nil {
		inner = replaceDCField(inner, "language", strSliceIfNotEmpty(*meta.Language))
	}
	if meta.Subjects != nil {
		inner = replaceDCField(inner, "subject", *meta.Subjects)
	}

	return []byte(before + openTag + inner + closeTag + after), nil
}

// replaceDCField removes all existing dc:name (or bare name) elements from
// the metadata inner content and appends new <dc:name> elements for each value.
// DC elements only ever contain plain text, so a text-only regex is safe here.
func replaceDCField(inner, name string, values []string) string {
	// re matches a Dublin Core element, capturing style and tags.
	// 1: Indentation/prefix whitespace
	// 2: Opening tag name (including namespace prefix if any)
	// 3: Attributes (including leading space, excluding trailing >)
	// 4: Closing tag
	// 5: Trailing whitespace/newline
	re := regexp.MustCompile(
		`(?i)([ \t]*)<((?:[a-zA-Z0-9_-]+:)?` + regexp.QuoteMeta(name) + `)(\s[^>]*)?>` +
			`[^<]*` +
			`(</(?:[a-zA-Z0-9_-]+:)?` + regexp.QuoteMeta(name) + `>)([ \t]*\n?)`,
	)

	matches := re.FindAllStringSubmatchIndex(inner, -1)

	var cleanValues []string
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			cleanValues = append(cleanValues, t)
		}
	}

	if len(matches) == 0 {
		if len(cleanValues) == 0 {
			return inner
		}
		// No original tags found; append new ones at the end.
		var sb strings.Builder
		for _, v := range cleanValues {
			sb.WriteString("\n    <dc:")
			sb.WriteString(name)
			sb.WriteString(">")
			xmlEscapeText(&sb, v)
			sb.WriteString("</dc:")
			sb.WriteString(name)
			sb.WriteString(">")
		}
		return inner + sb.String()
	}

	var sb strings.Builder
	lastEnd := 0

	for i, m := range matches {
		sb.WriteString(inner[lastEnd:m[0]]) // Preserve content before/between tags

		if i < len(cleanValues) {
			// Update existing tag, preserving attributes and prefix.
			sb.WriteString(inner[m[2]:m[3]]) // Indent
			sb.WriteString("<")
			sb.WriteString(inner[m[4]:m[5]]) // Name
			if m[6] != -1 {
				sb.WriteString(inner[m[6]:m[7]]) // Attributes
			}
			sb.WriteString(">")
			xmlEscapeText(&sb, cleanValues[i])
			sb.WriteString(inner[m[8]:m[9]])   // Closing tag
			sb.WriteString(inner[m[10]:m[11]]) // Trailing whitespace
		}
		// If i >= len(cleanValues), the match is skipped (removed).

		lastEnd = m[1]

		// If this was the last original tag but we have more values to write,
		// append them immediately after it, copying the style of the last match.
		if i == len(matches)-1 && len(cleanValues) > len(matches) {
			indent := inner[m[2]:m[3]]
			nameWithPrefix := inner[m[4]:m[5]]
			closeTag := inner[m[8]:m[9]]
			newline := inner[m[10]:m[11]]
			if newline == "" {
				newline = "\n"
			}

			for j := len(matches); j < len(cleanValues); j++ {
				sb.WriteString(indent)
				sb.WriteString("<")
				sb.WriteString(nameWithPrefix)
				sb.WriteString(">")
				xmlEscapeText(&sb, cleanValues[j])
				sb.WriteString(closeTag)
				sb.WriteString(newline)
			}
		}
	}
	sb.WriteString(inner[lastEnd:]) // Preserve remaining content

	return sb.String()
}

func xmlEscapeText(sb *strings.Builder, s string) {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	sb.Write(buf.Bytes())
}

func strSliceIfNotEmpty(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return []string{s}
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	var out []string
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
