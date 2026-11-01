package site

import (
	"bytes"
	"maps"
	"mime"
	"strings"
	"testing"
)

func TestHomeOffersJXLOnlyWhenBuilt(t *testing.T) {
	saved := etags
	t.Cleanup(func() { etags = saved })

	render := func() string {
		var buf bytes.Buffer
		if err := templates.ExecuteTemplate(&buf, "home.html", yearData()); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	sources := []string{
		`<source srcset="/images/me.jxl" type="image/jxl" />`,
		`<source srcset="images/trees.jxl" type="image/jxl" />`,
	}
	fallbacks := []string{`src="/images/me.jpg"`, `src="images/trees.jpg"`}

	// A plain `go run`: only the JPEGs are embedded.
	etags = maps.Clone(saved)
	delete(etags, "images/me.jxl")
	delete(etags, "images/trees.jxl")
	page := render()
	for _, s := range sources {
		if strings.Contains(page, s) {
			t.Errorf("without .jxl files, page still offers %s", s)
		}
	}

	// A Docker build: the image stage added .jxl files.
	etags["images/me.jxl"] = `"test"`
	etags["images/trees.jxl"] = `"test"`
	page = render()
	for _, s := range append(sources, fallbacks...) {
		if !strings.Contains(page, s) {
			t.Errorf("with .jxl files, page is missing %s", s)
		}
	}
}

func TestJXLContentType(t *testing.T) {
	if got := mime.TypeByExtension(".jxl"); got != "image/jxl" {
		t.Errorf("Content-Type for .jxl = %q, want image/jxl", got)
	}
}
