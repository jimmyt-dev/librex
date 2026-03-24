package metadata

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path"
	"path/filepath"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

// BookMeta holds Dublin Core metadata and cover image extracted from a book file.
type BookMeta struct {
	Title       string
	Creator     string // dc:creator (primary author)
	Subject     string // dc:subject (comma-separated topics/genres)
	Description string
	Publisher   string
	Contributor string // dc:contributor
	Date        string // dc:date (publication date)
	Type        string // dc:type
	Format      string // dc:format (MIME type)
	Identifier  string // dc:identifier (ISBN, URI, etc.)
	Source      string // dc:source
	Language    string
	Relation    string // dc:relation
	Coverage    string // dc:coverage

	CoverImage []byte // raw cover image bytes (nil if not found)
	CoverMime  string // e.g. "image/jpeg"
}

// Extract reads metadata from a book file based on its extension.
func Extract(filePath string) BookMeta {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub":
		return extractEPUB(filePath)
	case ".pdf":
		return extractPDF(filePath)
	default:
		return BookMeta{}
	}
}

// --- EPUB (Dublin Core from OPF) ---

type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
	Manifest opfManifest `xml:"manifest"`
}

type opfMetadata struct {
	Title       []string  `xml:"title"`
	Creator     []string  `xml:"creator"`
	Subject     []string  `xml:"subject"`
	Description []string  `xml:"description"`
	Publisher   []string  `xml:"publisher"`
	Contributor []string  `xml:"contributor"`
	Date        []string  `xml:"date"`
	Type        []string  `xml:"type"`
	Format      []string  `xml:"format"`
	Identifier  []string  `xml:"identifier"`
	Source      []string  `xml:"source"`
	Language    []string  `xml:"language"`
	Relation    []string  `xml:"relation"`
	Coverage    []string  `xml:"coverage"`
	Meta        []opfMeta `xml:"meta"`
}

type opfMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type opfManifest struct {
	Items []manifestItem `xml:"item"`
}

type manifestItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type container struct {
	XMLName  xml.Name   `xml:"container"`
	Rootfile []rootfile `xml:"rootfiles>rootfile"`
}

type rootfile struct {
	FullPath string `xml:"full-path,attr"`
}

func first(ss []string) string {
	for _, s := range ss {
		if v := strings.TrimSpace(s); v != "" {
			return v
		}
	}
	return ""
}

func filterEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if v := strings.TrimSpace(s); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func joinAll(ss []string) string {
	var out []string
	for _, s := range ss {
		if v := strings.TrimSpace(s); v != "" {
			out = append(out, v)
		}
	}
	return strings.Join(out, ", ")
}

func extractEPUB(filePath string) BookMeta {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return BookMeta{}
	}
	defer r.Close()

	// Build a quick name→file index for the zip
	fileIndex := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		fileIndex[f.Name] = f
	}

	// Find OPF path from META-INF/container.xml
	opfPath := ""
	for name := range fileIndex {
		if strings.EqualFold(name, "META-INF/container.xml") {
			rc, err := fileIndex[name].Open()
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
		return BookMeta{}
	}

	opfFile, ok := fileIndex[opfPath]
	if !ok {
		return BookMeta{}
	}

	rc, err := opfFile.Open()
	if err != nil {
		return BookMeta{}
	}
	data, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return BookMeta{}
	}

	var pkg opfPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return BookMeta{}
	}

	m := pkg.Metadata
	meta := BookMeta{
		Title:       first(m.Title),
		Creator:     first(m.Creator),
		Subject:     joinAll(m.Subject),
		Description: first(m.Description),
		Publisher:   first(m.Publisher),
		Contributor: joinAll(m.Contributor),
		Date:        first(m.Date),
		Type:        first(m.Type),
		Format:      first(m.Format),
		Identifier:  first(m.Identifier),
		Source:      first(m.Source),
		Language:    first(m.Language),
		Relation:    first(m.Relation),
		Coverage:    first(m.Coverage),
	}

	// Find cover image in manifest
	// EPUB3: item with properties="cover-image"
	// EPUB2: <meta name="cover" content="item-id"/> pointing to manifest item
	coverID := ""
	for _, mt := range m.Meta {
		if strings.EqualFold(mt.Name, "cover") && mt.Content != "" {
			coverID = mt.Content
			break
		}
	}

	opfDir := path.Dir(opfPath)
	if opfDir == "." {
		opfDir = ""
	}

	for _, item := range pkg.Manifest.Items {
		isCover := strings.Contains(item.Properties, "cover-image") ||
			(coverID != "" && item.ID == coverID) ||
			(coverID == "" && (strings.EqualFold(item.ID, "cover-image") || strings.EqualFold(item.ID, "cover")))

		if !isCover || !strings.HasPrefix(item.MediaType, "image/") {
			continue
		}

		// Resolve href relative to OPF directory
		imgPath := item.Href
		if opfDir != "" {
			imgPath = opfDir + "/" + item.Href
		}
		imgPath = path.Clean(imgPath)

		imgFile, ok := fileIndex[imgPath]
		if !ok {
			continue
		}
		irc, err := imgFile.Open()
		if err != nil {
			continue
		}
		imgBytes, err := io.ReadAll(irc)
		irc.Close()
		if err != nil {
			continue
		}

		meta.CoverImage = imgBytes
		meta.CoverMime = item.MediaType
		break
	}

	return meta
}

// --- PDF ---

func extractPDF(filePath string) BookMeta {
	ctx, err := pdfcpuapi.ReadContextFile(filePath)
	if err != nil {
		return BookMeta{}
	}

	return BookMeta{
		Title:   ctx.Title,
		Creator: ctx.Author,
		Subject: strings.Join(filterEmpty(ctx.Subject, ctx.Keywords), ", "),
	}
}
