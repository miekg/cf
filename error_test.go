package cf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	dir := "testdata/error"
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
		buf, err := os.ReadFile(dir + "/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		errortext, _ := os.ReadFile(dir + "/" + f.Name() + ".error")

		t.Run(f.Name(), func(t *testing.T) {
			tokens, err := Lex(string(buf))
			if err != nil {
				t.Fatal(err)
			}

			_, _, err = ParseTokens(tokens)
			if err == nil {
				t.Fatal("expected error, got none")
			}
			err1 := strings.TrimSpace(err.Error()) + "\n"
			if err1 != string(errortext) {
				t.Fatalf("expected %q, got %q", errortext, err1)
			}
		})
	}
}
