package cf

//go:generate goyacc -v "" parse.y

import (
	"fmt"
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
