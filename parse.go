// Package cf is used to parse CFEngine .cf files and convert them into an AST. With the Print function(s) this AST can
// be pretty printed. Think as it as a gofmt for CFEngine.
//
// Not all syntax is parsed correctly. Currently:
//
// - Comments that are placed at the end of a bundle/body are silently dropped.
// - Multiline comments with escaped quoting characters will lead to a lexer error.
// - Macros (@if etc) are not parsed (lexer error)
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

// Some helper functions for parser usage. Mostly to make code smaller.
func printf(format string, v ...interface{})    { fmt.Printf(format, v...) }
func debug(y yyLexer, s string, t ...ast.Token) { y.(*Lexer).yydebug(s, t...) }
func p(y yyLexer) ast.Node                      { return y.(*Lexer).parent }
func setP(y yyLexer, p ast.Node)                { y.(*Lexer).parent = p }

// Parse parses a CFengine file using the lexer l and returns the AST. The parser is not concurrent safe, but can be re-used.
func Parse(l *Lexer) (ast.Node, error) {
	yyParse(l)
	return l.Spec, l.Err
}
