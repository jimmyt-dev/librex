package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"reliquary/internal/db"
)

// PatternData holds the variables available for file naming pattern resolution.
type PatternData struct {
	Authors     string // primary author or "Unknown"
	Title       string
	Series      string // empty if no series
	SeriesIndex string // zero-padded series number, empty if none
	Year        string // extracted from date field
	Publisher   string
	Language    string
	Ext         string // file extension including dot, e.g. ".epub"
}

// resolveFilePattern takes a pattern template and data, returns a relative file path.
//
// Syntax:
//   - {var} — replaced with value, or "Unknown" if empty
//   - <text {var} text> — entire segment omitted if {var} is empty
//   - Literal text is preserved as-is
func resolveFilePattern(pattern string, data PatternData) string {
	vars := map[string]string{
		"authors":     data.Authors,
		"title":       data.Title,
		"series":      data.Series,
		"seriesIndex": data.SeriesIndex,
		"year":        data.Year,
		"publisher":   data.Publisher,
		"language":    data.Language,
		"ext":         data.Ext,
	}

	var result strings.Builder
	i := 0
	for i < len(pattern) {
		if pattern[i] == '<' {
			end := strings.IndexByte(pattern[i+1:], '>')
			if end == -1 {
				result.WriteByte(pattern[i])
				i++
				continue
			}
			segment := pattern[i+1 : i+1+end]
			resolved, hasEmpty := resolveSegment(segment, vars)
			if !hasEmpty {
				result.WriteString(resolved)
			}
			i = i + 1 + end + 1
		} else if pattern[i] == '{' {
			end := strings.IndexByte(pattern[i+1:], '}')
			if end == -1 {
				result.WriteByte(pattern[i])
				i++
				continue
			}
			varName := pattern[i+1 : i+1+end]
			if val, ok := vars[varName]; ok && val != "" {
				result.WriteString(val)
			} else if varName == "ext" {
				result.WriteString(data.Ext)
			} else {
				result.WriteString("Unknown")
			}
			i = i + 1 + end + 1
		} else {
			result.WriteByte(pattern[i])
			i++
		}
	}

	// Sanitize each path segment individually
	raw := result.String()
	parts := strings.Split(raw, "/")
	for j, p := range parts {
		parts[j] = sanitizeName(p)
	}

	return filepath.Join(parts...)
}

// resolveSegment resolves {var} references within an optional <...> segment.
// Returns the resolved string and whether any variable was empty.
func resolveSegment(segment string, vars map[string]string) (string, bool) {
	var result strings.Builder
	hasEmpty := false
	i := 0
	for i < len(segment) {
		if segment[i] == '{' {
			end := strings.IndexByte(segment[i+1:], '}')
			if end == -1 {
				result.WriteByte(segment[i])
				i++
				continue
			}
			varName := segment[i+1 : i+1+end]
			if val, ok := vars[varName]; ok && val != "" {
				result.WriteString(val)
			} else {
				hasEmpty = true
			}
			i = i + 1 + end + 1
		} else {
			result.WriteByte(segment[i])
			i++
		}
	}
	return result.String(), hasEmpty
}

// buildPatternData constructs PatternData from staged book fields.
func buildPatternData(title string, author *string, date *string, publisher *string, language *string, seriesName *string, seriesNumber *float64, ext string) PatternData {
	d := PatternData{
		Authors: "Unknown",
		Title:   title,
		Ext:     ext,
	}
	if author != nil && *author != "" {
		names := parseAuthorString(*author)
		if len(names) > 0 {
			d.Authors = names[0]
		}
	}
	if date != nil && *date != "" {
		d.Year = extractYear(*date)
	}
	if publisher != nil && *publisher != "" {
		d.Publisher = *publisher
	}
	if language != nil && *language != "" {
		d.Language = *language
	}
	if seriesName != nil && *seriesName != "" {
		d.Series = *seriesName
	}
	if seriesNumber != nil && *seriesNumber != 0 {
		d.SeriesIndex = formatSeriesIndex(*seriesNumber)
	}
	return d
}

// extractYear pulls a 4-digit year from a date string.
func extractYear(date string) string {
	for i := 0; i <= len(date)-4; i++ {
		if date[i] >= '0' && date[i] <= '9' &&
			date[i+1] >= '0' && date[i+1] <= '9' &&
			date[i+2] >= '0' && date[i+2] <= '9' &&
			date[i+3] >= '0' && date[i+3] <= '9' {
			year := date[i : i+4]
			if year >= "1000" && year <= "2999" {
				return year
			}
		}
	}
	return ""
}

// formatSeriesIndex zero-pads a series number.
func formatSeriesIndex(num float64) string {
	if num == float64(int(num)) {
		return fmt.Sprintf("%02d", int(num))
	}
	return fmt.Sprintf("%05.2f", num)
}

// getEffectivePattern returns the naming pattern to use: library override > user default > hardcoded default.
func getEffectivePattern(r *http.Request, libraryID, userID string) string {
	// Check library override first
	var libPattern *string
	_ = db.DB.QueryRow(r.Context(),
		"SELECT file_naming_pattern FROM libraries WHERE id = $1",
		libraryID).Scan(&libPattern)
	if libPattern != nil && *libPattern != "" {
		return *libPattern
	}

	// Fall back to user setting
	var userPattern string
	err := db.DB.QueryRow(r.Context(),
		"SELECT file_naming_pattern FROM user_settings WHERE user_id = $1",
		userID).Scan(&userPattern)
	if err == nil && userPattern != "" {
		return userPattern
	}

	return defaultFileNamingPattern
}
