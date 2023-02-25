package cf

//go:generate goyacc -v "" parse.y

import (
	"fmt"
	"os"

	"github.com/miekg/cf/ast"
)

func (l *Lexer) yydebug(s string, t ...ast.Token) {
	if !l.D {
		return
	}
	lit := ""
	if len(t) > 0 {
		lit = t[0].Lit
	}
	fmt.Fprintf(os.Stderr, "yy : token [%s] %q\n", lit, s) // align with lex debug
}

// Parse parses a CFengine file in r and returns the AST. The parser is not concurrent safe.
func Parse(l *Lexer) ast.Node {
	yyParse(l)
	return l.Spec
}
