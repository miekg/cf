package cf

import (
	"io"

	"github.com/miekg/cf/internal/parse"
	"github.com/shivamMg/rd"
)

// Parse parses the CFEngine file in buffer into an CFEngine AST. See ParseTokens().
func Parse(buffer string) (tree *rd.Tree, debugTree *rd.DebugTree, err error) {
	tokens, err := Lex(buffer)
	if err != nil {
		return nil, nil, err
	}

	return ParseTokens(tokens)
}

// Print pretty prints the CFengine AST in tree.
func Print(w io.Writer, tree *rd.Tree) { parse.Print(w, tree) }

// Lex returns the tokens from the CFEngine file in the buffer.
func Lex(buffer string) ([]rd.Token, error) { return parse.Lex(buffer) }

// ParseTokens parses the tokens in an CFEngine AST. An empty files returns no trees, but also no error.
func ParseTokens(tokens []rd.Token) (tree *rd.Tree, debugTree *rd.DebugTree, err error) {
	if len(tokens) == 0 {
		return nil, nil, nil
	}
	b := rd.NewBuilder(tokens)
	if ok := parse.Specification(b); !ok {
		return nil, b.DebugTree(), b.Err()
	}
	return b.ParseTree(), b.DebugTree(), nil
}
