package metadata

import (
	"strings"
	"testing"
)

// helpers
func strPtr(s string) *string    { return &s }
func slicePtr(ss []string) *[]string { return &ss }

// A realistic EPUB2 OPF with namespaces, attributes on dc:creator, spine, and guide.
// Tests must prove none of the non-metadata content is disturbed.
const epub2OPF = `<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>Original Title</dc:title>
    <dc:creator opf:role="aut" opf:file-as="Author, Test">Test Author</dc:creator>
    <dc:publisher>Original Publisher</dc:publisher>
    <dc:description>Original Description</dc:description>
    <dc:date>2020</dc:date>
    <dc:language>en</dc:language>
    <dc:subject>Fantasy</dc:subject>
    <dc:identifier id="book-id">urn:isbn:1234567890</dc:identifier>
    <meta name="cover" content="cover-image"/>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="cover-image" href="images/cover.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="chapter1"/>
    <itemref idref="chapter2"/>
  </spine>
  <guide>
    <reference type="cover" title="Cover" href="cover.html"/>
  </guide>
</package>`

func patch(t *testing.T, opf string, meta WriteMeta) string {
	t.Helper()
	out, err := patchOPFMetadata([]byte(opf), meta)
	if err != nil {
		t.Fatalf("patchOPFMetadata error: %v", err)
	}
	return string(out)
}

func assertContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected to find %q in output", msg, needle)
	}
}

func assertNotContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: expected NOT to find %q in output", msg, needle)
	}
}

// --- Structure preservation ---

func TestPatchOPF_PreservesSpine(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `<spine toc="ncx">`, "spine element")
	assertContains(t, out, `<itemref idref="chapter1"/>`, "spine itemref")
	assertContains(t, out, `<itemref idref="chapter2"/>`, "spine itemref 2")
}

func TestPatchOPF_PreservesManifest(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `<manifest>`, "manifest")
	assertContains(t, out, `media-type="application/x-dtbncx+xml"`, "manifest ncx item")
}

func TestPatchOPF_PreservesGuide(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `<guide>`, "guide element")
	assertContains(t, out, `type="cover"`, "guide reference")
}

func TestPatchOPF_PreservesPackageAttributes(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `version="2.0"`, "package version attr")
	assertContains(t, out, `unique-identifier="book-id"`, "unique-identifier attr")
}

func TestPatchOPF_PreservesIdentifier(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `urn:isbn:1234567890`, "dc:identifier untouched")
}

func TestPatchOPF_PreservesCoverMeta(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `<meta name="cover" content="cover-image"/>`, "cover meta untouched")
}

// --- Field updates ---

func TestPatchOPF_UpdateTitle(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"})
	assertContains(t, out, `<dc:title>New Title</dc:title>`, "new title")
	assertNotContains(t, out, "Original Title", "old title removed")
}

func TestPatchOPF_UpdatePublisher(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Publisher: strPtr("Seven Seas Entertainment")})
	assertContains(t, out, `<dc:publisher>Seven Seas Entertainment</dc:publisher>`, "new publisher")
	assertNotContains(t, out, "Original Publisher", "old publisher removed")
}

func TestPatchOPF_UpdateMultipleAuthors(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Authors: slicePtr([]string{"Author One", "Author Two"})})
	assertContains(t, out, `<dc:creator>Author One</dc:creator>`, "author one")
	assertContains(t, out, `<dc:creator>Author Two</dc:creator>`, "author two")
	assertNotContains(t, out, "Test Author", "old author removed")
}

func TestPatchOPF_UpdateSeriesViaSubjects(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Subjects: slicePtr([]string{"Action", "Romance"})})
	assertContains(t, out, `<dc:subject>Action</dc:subject>`, "subject action")
	assertContains(t, out, `<dc:subject>Romance</dc:subject>`, "subject romance")
	assertNotContains(t, out, ">Fantasy<", "old subject removed")
}

func TestPatchOPF_UpdateDescription(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Description: strPtr("A new description.")})
	assertContains(t, out, `<dc:description>A new description.</dc:description>`, "new description")
	assertNotContains(t, out, "Original Description", "old description removed")
}

func TestPatchOPF_UpdateLanguage(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Language: strPtr("ja")})
	assertContains(t, out, `<dc:language>ja</dc:language>`, "new language")
	assertNotContains(t, out, ">en<", "old language removed")
}

func TestPatchOPF_UpdateDate(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Date: strPtr("2024")})
	assertContains(t, out, `<dc:date>2024</dc:date>`, "new date")
	assertNotContains(t, out, ">2020<", "old date removed")
}

// --- Nil = don't touch ---

func TestPatchOPF_NilPublisherUntouched(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"}) // Publisher is nil
	assertContains(t, out, "Original Publisher", "publisher untouched when nil")
}

func TestPatchOPF_NilAuthorsUntouched(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"}) // Authors is nil
	assertContains(t, out, "Test Author", "authors untouched when nil")
}

func TestPatchOPF_NilSubjectsUntouched(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: "New Title"}) // Subjects is nil
	assertContains(t, out, "Fantasy", "subjects untouched when nil")
}

func TestPatchOPF_EmptyTitleUntouched(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Publisher: strPtr("New Pub")}) // Title ""
	assertContains(t, out, "Original Title", "title untouched when empty")
}

// --- Remove via empty pointer ---

func TestPatchOPF_RemovePublisher(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Publisher: strPtr("")})
	assertNotContains(t, out, "Original Publisher", "publisher removed")
	assertNotContains(t, out, "dc:publisher", "dc:publisher element gone")
}

func TestPatchOPF_RemoveDescription(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Description: strPtr("")})
	assertNotContains(t, out, "Original Description", "description removed")
}

func TestPatchOPF_RemoveAllAuthors(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Authors: slicePtr([]string{})})
	assertNotContains(t, out, "Test Author", "authors removed")
	assertNotContains(t, out, "dc:creator", "dc:creator element gone")
}

// --- XML escaping ---

func TestPatchOPF_EscapesSpecialChars(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{Title: `It's a "Test" & More`})
	assertContains(t, out, `It&#39;s a &#34;Test&#34; &amp; More`, "special chars escaped")
}

// --- Multiple updates in one call ---

func TestPatchOPF_MultipleFieldsAtOnce(t *testing.T) {
	out := patch(t, epub2OPF, WriteMeta{
		Title:       "New Title",
		Authors:     slicePtr([]string{"New Author"}),
		Publisher:   strPtr("New Publisher"),
		Description: strPtr("New Description"),
		Language:    strPtr("fr"),
		Date:        strPtr("2025"),
		Subjects:    slicePtr([]string{"Sci-Fi"}),
	})
	assertContains(t, out, "New Title", "title")
	assertContains(t, out, "New Author", "author")
	assertContains(t, out, "New Publisher", "publisher")
	assertContains(t, out, "New Description", "description")
	assertContains(t, out, ">fr<", "language")
	assertContains(t, out, ">2025<", "date")
	assertContains(t, out, "Sci-Fi", "subject")
	// Structure still intact
	assertContains(t, out, `<spine`, "spine")
	assertContains(t, out, `urn:isbn:1234567890`, "identifier")
}

// --- EPUB3 format (opf: namespace prefix on metadata tag) ---

const epub3OPF = `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/" version="3.0" unique-identifier="uid">
  <metadata>
    <dc:identifier id="uid">urn:uuid:abc123</dc:identifier>
    <dc:title>EPUB3 Book</dc:title>
    <dc:creator>Original Creator</dc:creator>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="toc" href="toc.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="toc"/>
  </spine>
</package>`

func TestPatchOPF_EPUB3_UpdateTitle(t *testing.T) {
	out := patch(t, epub3OPF, WriteMeta{Title: "New EPUB3 Title"})
	assertContains(t, out, "New EPUB3 Title", "new title")
	assertNotContains(t, out, "EPUB3 Book", "old title gone")
}

func TestPatchOPF_EPUB3_PreservesModifiedMeta(t *testing.T) {
	out := patch(t, epub3OPF, WriteMeta{Title: "New"})
	assertContains(t, out, `property="dcterms:modified"`, "EPUB3 meta preserved")
}

func TestPatchOPF_EPUB3_NilCreatorUntouched(t *testing.T) {
	out := patch(t, epub3OPF, WriteMeta{Title: "New"}) // Authors nil
	assertContains(t, out, "Original Creator", "creator untouched")
}

// --- Error cases ---

func TestPatchOPF_ErrorNoMetadataTag(t *testing.T) {
	_, err := patchOPFMetadata([]byte(`<package><manifest/></package>`), WriteMeta{Title: "X"})
	if err == nil {
		t.Error("expected error for OPF with no metadata element")
	}
}
