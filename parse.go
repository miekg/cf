package cf

//go:generate goyacc -v "" parse.y

import (
	"fmt"
	"io"
	"os"

	"github.com/miekg/cf/ast"
)

func (l *Lexer) yydebug(s string, t ...ast.Token) {
	if !l.d {
		return
	}
	lit := ""
	if len(t) > 0 {
		lit = t[0].Lit
	}
	fmt.Fprintf(os.Stderr, "yy : token [%s] %q\n", lit, s) // align with lex debug
}

// Parse parses a CFengine file in r and returns the AST.
func Parse(r io.Reader) ast.Node {
	l := NewLexer(r, false)
	yyParse(l)
	return l.Spec
}
