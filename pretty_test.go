package cf

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestPrettyPrint matches a pretty printed .cf file to a .pretty file to see if they match.
// Don't add to many pretty files, annoying to test
func TestPrettyPrint(t *testing.T) {
	dir := "testdata"
	testFiles, err := os.ReadDir(dir)
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
		buf, err := os.ReadFile("testdata/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		pretty, err := os.ReadFile("testdata/" + f.Name() + ".pretty")
		if err != nil {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			tree, _, err := Parse(string(buf))
			if err != nil {
				log.Fatal(err)
			}
			r := &bytes.Buffer{}
			Print(r, tree)
			tr := trimSpace(r.Bytes())
			tp := trimSpace(pretty)

			if string(tr) != string(tp) {
				t.Errorf("pretty and input, don't match")
			}
		})
	}
}
func trimSpace(buf []byte) []byte {
	buf = bytes.ReplaceAll(buf, []byte{' '}, nil)
	buf = bytes.ReplaceAll(buf, []byte{'\n'}, nil)
	buf = bytes.ReplaceAll(buf, []byte{'\t'}, nil)
	return buf
}
