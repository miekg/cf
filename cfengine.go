package cf

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/parse"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
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

// ParseTokens parses the tokens in an CFEngine AST. An empty file returns no trees, but also no error. When parsing
// fails an error is returned and the debugTree show what was successfully parsed and what wasn't.
func ParseTokens(tokens []rd.Token) (tree *rd.Tree, debugTree *rd.DebugTree, err error) {
	if len(tokens) == 0 {
		return nil, nil, nil
	}
	b := rd.NewBuilder(tokens)
	if ok := parse.Specification(b); !ok {
		err = fmt.Errorf("parsing error around token %q on line %d", b.ErrorToken.(token.T).Value, b.ErrorToken.(token.T).Line)
		return nil, b.DebugTree(), err
	}
	return b.ParseTree(), b.DebugTree(), nil
}

// IsNoParse returns true if the first token in tokens is a comment and contains '# cffmt:no' isNoParse returns true.
func IsNoParse(tokens []rd.Token) bool {
	if len(tokens) == 0 {
		return false
	}

	if ct, ok := tokens[0].(token.T); ok {
		if ct.Type == chroma.Comment && strings.HasPrefix(ct.Value, "# cffmt:no") {
			return true
		}
	}
	return false
}
