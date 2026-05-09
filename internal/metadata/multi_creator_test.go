package metadata

import (
	"strings"
	"testing"
)

func TestPatchOPF_EPUB3MultiCreatorRemoveAll(t *testing.T) {
	opf := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/" version="3.0">
  <metadata>
    <dc:title id="id">In the Land of Leadale, Vol. 5</dc:title>
    <dc:creator id="id-1">Ceez</dc:creator>
    <dc:creator id="id-2">Tenmaso</dc:creator>
    <dc:language>en</dc:language>
    <meta refines="#id-1" property="role" scheme="marc:relators">aut</meta>
    <meta refines="#id-1" property="file-as">Ceez</meta>
    <meta refines="#id-2" property="role" scheme="marc:relators">aut</meta>
    <meta refines="#id-2" property="file-as">Tenmaso</meta>
  </metadata>
</package>`

	out := patch(t, opf, WriteMeta{Authors: &[]string{}})
	if strings.Contains(out, "dc:creator") {
		t.Errorf("expected no dc:creator elements, got:\n%s", out)
	}
	if strings.Contains(out, "Ceez") || strings.Contains(out, "Tenmaso") {
		t.Errorf("expected no author names, got:\n%s", out)
	}
}

func TestPatchOPF_EPUB3MultiCreatorSetAuthors(t *testing.T) {
	opf := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/" version="3.0">
  <metadata>
    <dc:title id="id">In the Land of Leadale, Vol. 5</dc:title>
    <dc:creator id="id-1">Ceez</dc:creator>
    <dc:creator id="id-2">Tenmaso</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
</package>`

	out := patch(t, opf, WriteMeta{Authors: &[]string{"Ceez", "Tenmaso"}})
	if !strings.Contains(out, "Ceez") || !strings.Contains(out, "Tenmaso") {
		t.Errorf("expected both authors, got:\n%s", out)
	}
}
