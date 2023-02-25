package cf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/miekg/cf/ast"
)

// Symbol is used to construct the regular expressions used in the lexer.
type Symbol struct {
	tok int
	exp *regexp.Regexp
}

const (
	DONE    = iota // not in cfengine
	NONE           // not in cfengine
	COMMENT        // not in cfengine
	SPACE          // not in cfengine, skipped when lexed
	CHAR           // single character, converted to literal in lexer

	/* defined via yacc, %token in parse.y:
	BUNDLE
	BODY
	PROMISE
	NAKEDVAR
	IDENTIFIER
	FATARROW
	THINARROW
	QSTRING
	CLASSGUARD
	PROMISEGUARD
	*/
)

var SymbolText = map[int]string{
	DONE:    "",
	NONE:    "",
	COMMENT: "comment",
	CHAR:    "char",
	SPACE:   "space",

	BUNDLE:       "bundle",
	BODY:         "body",
	PROMISE:      "promise",
	NAKEDVAR:     "nakedvar",
	IDENTIFIER:   "identifier",
	FATARROW:     "fatarrow",
	THINARROW:    "thinarrow",
	QSTRING:      "qstring",
	CLASSGUARD:   "classguard",
	PROMISEGUARD: "promiseguard",
}

// Symbols for cfengine, order of list taken from cf3lex.l, excluding 'space'
var Symbols = []Symbol{bundle, body, promise, identifier, symbol, fatarrow, thinarrow, varclass, class, promiseguard,
	qstringquote, qstringsquote, qstringbacktick, nakedvar, comment, char}

// from: cfengine/core/libpromises/cf3lex.l
var (
	comment         = Symbol{COMMENT, regexp.MustCompilePOSIX(`^#[^\n]*`)}
	bundle          = Symbol{BUNDLE, regexp.MustCompilePOSIX(`^bundle`)}
	body            = Symbol{BODY, regexp.MustCompilePOSIX(`^body`)}
	promise         = Symbol{PROMISE, regexp.MustCompilePOSIX(`^promise`)}
	nakedvar        = Symbol{NAKEDVAR, regexp.MustCompilePOSIX(`^[$@][(][a-zA-Z0-9_\[\]\200-\377.:]+[)]|^[$@][{][a-zA-Z0-9_\[\]\200-\377.:]+[}]|^[$@][(][a-zA-Z0-9_\200-\377.:]+[\[][a-zA-Z0-9_$(){}\200-\377.:]+[\]]+[)]|^[$@][{][a-zA-Z0-9_\200-\377.:]+[\[][a-zA-Z0-9_$(){}\200-\377.:]+[\]]+[}]`)}
	identifier      = Symbol{IDENTIFIER, regexp.MustCompilePOSIX(`^[a-zA-Z0-9_]+`)}
	symbol          = Symbol{IDENTIFIER, regexp.MustCompilePOSIX(`^[a-zA-Z0-9_\200-\377]+[:][a-zA-Z0-9_\200-\377]+`)}
	fatarrow        = Symbol{FATARROW, regexp.MustCompilePOSIX(`^=>`)}
	thinarrow       = Symbol{THINARROW, regexp.MustCompilePOSIX(`^->`)}
	class           = Symbol{CLASSGUARD, regexp.MustCompilePOSIX(`^[.|&!()a-zA-Z0-9_\200-\377:][\t .|&!()a-zA-Z0-9_\200-\377:]*::`)}
	varclass        = Symbol{CLASSGUARD, regexp.MustCompilePOSIX(`^(\"[^"\0]*\"|\'[^'\0]*\')::`)}
	promiseguard    = Symbol{PROMISEGUARD, regexp.MustCompilePOSIX(`^[a-zA-Z_]+:`)}
	qstringsquote   = Symbol{QSTRING, regexp.MustCompilePOSIX(`^'[^']*'`)} // doesn't do the backspacing, of quote symbol
	qstringquote    = Symbol{QSTRING, regexp.MustCompilePOSIX(`^"[^"]*"`)}
	qstringbacktick = Symbol{QSTRING, regexp.MustCompilePOSIX("^`[^`]*`")}
	char            = Symbol{CHAR, regexp.MustCompilePOSIX(`^.`)}
)

// Lexer is steered from yacc to deliver tokens.
type Lexer struct {
	buf []byte // leftover from last match, deplete first before scanning
	*bufio.Scanner
	symbols []Symbol
	d       bool

	Spec   ast.Node // AST of parsed document
	parent ast.Node
}

// NewLexer retursn a pointer to a usuable Lexer.
func NewLexer(r io.Reader, debug ...bool) *Lexer {
	s := bufio.NewScanner(r)
	s.Split(scanLines)
	d := false
	if len(debug) > 0 {
		d = debug[0]
	}
	return &Lexer{Scanner: s, symbols: Symbols, d: d, parent: ast.New(&ast.Specification{}, ast.Token{})}
}

// Implemented for goyacc.
func (l *Lexer) Lex(lval *yySymType) int {
	rem := []string{}
Rescan:
	t := l.scan()
	switch t.Tok {
	case COMMENT:
		// TODO(miek): either we hang comments on the previous token, or the next. Either way
		// we'll have a problem for comments at the beginning or end - assume end-of-file comments are not
		// important.
		rem = append(rem, t.Lit)
		goto Rescan
	case SPACE:
		// skip
		goto Rescan
	default:
		t.Comment = rem
	}

	l.debug(t)
	lval.token = t
	return t.Tok
}

// Implemented for goyacc.
func (l *Lexer) Error(e string) {
	if len(l.buf) > 0 {
		fmt.Printf("error while parsing (left: %q): %s\n", l.buf, e)
		return
	}
	fmt.Printf("error while parsing %s\n", e)
}

func (l *Lexer) debug(t ast.Token) {
	if !l.d {
		return
	}
	st := SymbolText[t.Tok]
	if st == "" {
		st = t.Lit
	}
	fmt.Fprintf(os.Stderr, "lex: token [%s] %q\n", st, t.Lit)
}

func (l *Lexer) scan() ast.Token {
	if len(l.buf) == 0 {
		more := l.Scanner.Scan()
		if !more {
			return ast.Token{Tok: DONE, Lit: ""}
		}
		l.buf = l.Bytes()
	}

	max := 0
	t := ast.Token{Tok: SPACE, Lit: ""} // will be skipped when nothing matches, happens on newlines or empty lines
	for _, s := range Symbols {
		match := s.exp.Find(l.buf)
		if match == nil {
			continue
		}
		if len(match) > max {
			max = len(match)
			switch s.tok {
			case CHAR:
				lit := bytes.TrimSpace(match)
				if len(lit) == 0 { // hack around parse errors
					t = ast.Token{Tok: SPACE, Lit: string(lit)} // single literal character
				} else {
					t = ast.Token{Tok: int(match[0]), Lit: string(lit)} // single literal character
				}
			default:
				lit := bytes.TrimSpace(match)
				t = ast.Token{Tok: s.tok, Lit: string(lit)} // single literal character
			}
		}
	}
	l.buf = l.buf[max:]
	return t
}

// copied from std lib
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return bytes.TrimSpace(data[0 : len(data)-1])
	}
	if len(data) > 0 {
		return bytes.TrimSpace(data)
	}
	return data
}