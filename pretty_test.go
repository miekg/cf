package cf

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestPrettyPrint matches a pretty printed .cf file to a .pretty file to see if they match.
func TestPrettyPrint(t *testing.T) {
	dir := "testdata"
	testFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatalf("could not read %s: %q", dir, err)
	}
	for _, f := range testFiles {
		if f.IsDir() {
			continue
		}

		if filepath.Ext(f.Name()) != ".cf" {
			continue
		}

		r, err := os.Open("testdata/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("looking at testdata/%s", f.Name())

		l := NewLexer(r)
		spec, err := Parse(l)
		r.Close()
		if err != nil {
			t.Errorf("failed to parse document: %s", err)
			continue
		}

		doc := &bytes.Buffer{}
		Print(doc, spec)

		// check for .pretty file
		fp := f.Name()[:len(f.Name())-3] + ".pretty"
		pretty, err := os.ReadFile("testdata/" + fp)
		if err != nil {
			t.Logf("No .pretty file for %s", f.Name())
			continue
		}

		if doc.String() != string(pretty) {
			t.Errorf("Pretty print of %s, doesn't match", f.Name())
		}
	}
}
