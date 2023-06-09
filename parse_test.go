package cf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
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
		buf, err := os.ReadFile(dir + "/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		ast, _ := os.ReadFile(dir + "/" + f.Name() + ".ast")

		t.Run(f.Name(), func(t *testing.T) {
			tokens, err := Lex(string(buf))
			if err != nil {
				t.Fatal(err)
			}

			parseTree, debugTree, err := ParseTokens(tokens)
			if err != nil {
				for i := range tokens {
					t.Logf("%v\n", tokens[i])
				}
				t.Log("Debug Tree:\n\n", debugTree)
				t.Fatal(err)
			}
			if ast != nil {
				if err := astCompare(string(ast), parseTree.String()); err != nil {
					t.Errorf("Test %q, AST doesn't match: %s", f.Name(), err)
					t.Logf("Test %q, AST\n%s\n", f.Name(), parseTree.String())
					t.Logf("Expect AST\n%s\n", string(ast))
					t.Log("Debug tree:\n\n", debugTree)
				}
			}
		})
	}
}
