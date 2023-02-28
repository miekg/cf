// Package cf is used to parse CFEngine .cf files and convert them into an AST. With the Print function(s) this AST can
// be pretty printed. Think as it as a gofmt for CFEngine.
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

// Parse parses a CFengine file in r and returns the AST. The parser is not concurrent safe. But can be re-used.
func Parse(l *Lexer) (ast.Node, error) {
	yyParse(l)
	return l.Spec, l.Err
}
