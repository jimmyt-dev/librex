package metadata

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

type BookMeta struct {
	Title  string
	Author string
}

// Extract reads metadata from a book file based on its extension.
// Returns whatever metadata it can find; missing fields are empty strings.
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

// --- EPUB ---

type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
}

type opfMetadata struct {
	Title   []string `xml:"title"`
	Creator []string `xml:"creator"`
}

type container struct {
	XMLName  xml.Name  `xml:"container"`
	Rootfile []rootfile `xml:"rootfiles>rootfile"`
}

type rootfile struct {
	FullPath string `xml:"full-path,attr"`
}

func extractEPUB(path string) BookMeta {
	r, err := zip.OpenReader(path)
	if err != nil {
		return BookMeta{}
	}
	defer r.Close()

	// Find OPF path from META-INF/container.xml
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
		return BookMeta{}
	}

	// Parse OPF
	for _, f := range r.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				return BookMeta{}
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return BookMeta{}
			}

			var pkg opfPackage
			if err := xml.Unmarshal(data, &pkg); err != nil {
				return BookMeta{}
			}

			meta := BookMeta{}
			if len(pkg.Metadata.Title) > 0 {
				meta.Title = strings.TrimSpace(pkg.Metadata.Title[0])
			}
			if len(pkg.Metadata.Creator) > 0 {
				meta.Author = strings.TrimSpace(pkg.Metadata.Creator[0])
			}
			return meta
		}
	}

	return BookMeta{}
}

// --- PDF ---

func extractPDF(path string) BookMeta {
	ctx, err := pdfcpuapi.ReadContextFile(path)
	if err != nil {
		return BookMeta{}
	}

	meta := BookMeta{}
	if ctx.Title != "" {
		meta.Title = ctx.Title
	}
	if ctx.Author != "" {
		meta.Author = ctx.Author
	}
	return meta
}
