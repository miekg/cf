package cf

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
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
		yyParse(l)
		r.Close()

		doc := &bytes.Buffer{}
		Print(doc, l.Spec)
		dbuf := removeSpace(doc.Bytes())

		fbuf, _ := os.ReadFile("testdata/" + f.Name())
		fbuf = removeSpace(fbuf)

		if string(dbuf) != string(fbuf) {
			t.Errorf("file %s, pretty printed output is different from souce", f.Name())
			t.Logf("test with: wdiff -123 -s <(tr -d '[:space:]' < testdata/%s) <(cmd/cf/cf testdata/%s | tr -d '[:space:]')", f.Name(), f.Name())
			t.Logf("\n%s\n%s\n", dbuf, fbuf)
		}
	}
}

func removeSpace(buf []byte) []byte {
	// klunky, but good enough
	buf = bytes.ReplaceAll(buf, []byte{' '}, nil)
	buf = bytes.ReplaceAll(buf, []byte{'\n'}, nil)
	buf = bytes.ReplaceAll(buf, []byte{'\t'}, nil)
	return buf
}
