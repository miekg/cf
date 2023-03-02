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

// sym is used to construct the regular expressions used in the lexer.
type sym struct {
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

var symbolText = map[int]string{
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

// syms for cfengine, order of list taken from cf3lex.l, excluding 'space'
var syms = []sym{bundle, body, promise, identifier, symbol, fatarrow, thinarrow, varclass, class, promiseguard,
	qstringquote, qstringsquote, qstringbacktick, nakedvar, comment, char}

// from: cfengine/core/libpromises/cf3lex.l
var (
	comment      = sym{COMMENT, regexp.MustCompilePOSIX(`^#[^\n]*`)}
	bundle       = sym{BUNDLE, regexp.MustCompilePOSIX(`^bundle`)}
	body         = sym{BODY, regexp.MustCompilePOSIX(`^body`)}
	promise      = sym{PROMISE, regexp.MustCompilePOSIX(`^promise`)}
	nakedvar     = sym{NAKEDVAR, regexp.MustCompilePOSIX(`^[$@][(][a-zA-Z0-9_\[\]\200-\377.:]+[)]|^[$@][{][a-zA-Z0-9_\[\]\200-\377.:]+[}]|^[$@][(][a-zA-Z0-9_\200-\377.:]+[\[][a-zA-Z0-9_$(){}\200-\377.:]+[\]]+[)]|^[$@][{][a-zA-Z0-9_\200-\377.:]+[\[][a-zA-Z0-9_$(){}\200-\377.:]+[\]]+[}]`)}
	identifier   = sym{IDENTIFIER, regexp.MustCompilePOSIX(`^[a-zA-Z0-9_]+`)}
	symbol       = sym{IDENTIFIER, regexp.MustCompilePOSIX(`^[a-zA-Z0-9_\200-\377]+[:][a-zA-Z0-9_\200-\377]+`)}
	fatarrow     = sym{FATARROW, regexp.MustCompilePOSIX(`^=>`)}
	thinarrow    = sym{THINARROW, regexp.MustCompilePOSIX(`^->`)}
	class        = sym{CLASSGUARD, regexp.MustCompilePOSIX(`^[.|&!()a-zA-Z0-9_\200-\377:][\t .|&!()a-zA-Z0-9_\200-\377:]*::`)}
	varclass     = sym{CLASSGUARD, regexp.MustCompilePOSIX(`^(\"[^"\0]*\"|\'[^'\0]*\')::`)}
	promiseguard = sym{PROMISEGUARD, regexp.MustCompilePOSIX(`^[a-zA-Z_]+:`)}
	char         = sym{CHAR, regexp.MustCompilePOSIX(`^.`)}
	// original qstring regexp: \"((\\(.|\n))|[^"\\])*\"|\'((\\(.|\n))|[^'\\])*\'|`[^`]*`
	qstringsquote   = sym{QSTRING, regexp.MustCompilePOSIX(`^\'((\\(.|\n))|[^'\\])*\'`)}
	qstringquote    = sym{QSTRING, regexp.MustCompilePOSIX(`^\"((\\(.|\n))|[^"\\])*\"`)}
	qstringbacktick = sym{QSTRING, regexp.MustCompilePOSIX("^`[^`]*`")}
)

// Lexer is steered from yacc to deliver tokens.
type Lexer struct {
	buf  []byte // leftover from last match, deplete first before scanning
	name string // potential filename
	*bufio.Scanner
	symbols []sym
	parent  ast.Node

	D    bool     // If true enable debugging.
	Spec ast.Node // AST of parsed document.
	Err  error    // Set to the last error we see.

	col  int // position of token
	line int
}

// NewLexer returns a pointer to a usuable Lexer.
func NewLexer(r io.Reader, filename ...string) *Lexer {
	s := bufio.NewScanner(r)
	s.Split(scanLines)
	f := "<stdin>"
	if len(filename) > 0 {
		f = filename[0]
	}
	return &Lexer{Scanner: s, symbols: syms, D: false, name: f, parent: ast.New(&ast.Specification{}, ast.Token{})}
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
		// Hack to scan for multiline qstrings, currently only handles "-qstrings.
		// Scan until we see t.Lit == " or ' again, and restich the lines.
		multiline := ""
		switch q := t.Lit; q {
		case `"`, "'", "`":
			for t := l.scan(); t.Lit != q; t = l.scan() {
				switch t.Tok {
				case SPACE:
					multiline += " "
				case DONE:
					goto End
				default:
					multiline += t.Lit
				}
				if t.Newline {
					multiline += "\n"
				}
			}
			t.Lit = q + multiline + q
			t.Tok = QSTRING
		}

		t.Comment = rem
	}

End:
	l.debug(t)
	lval.token = t
	return t.Tok
}

// Implemented for goyacc.
func (l *Lexer) Error(e string) {
	if len(l.buf) > 0 {
		l.Err = fmt.Errorf("%s:%d:%d: error while parsing (left in buffer: %q): %s\n", l.name, l.line, l.col, l.buf, e)
		return
	}
	l.Err = fmt.Errorf("%s:%d:%d error while parsing: %s\n", l.name, l.line, l.col, e)
}

func (l *Lexer) debug(t ast.Token) {
	if !l.D {
		return
	}
	st := symbolText[t.Tok]
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
		l.line++
		l.col = 0
		l.buf = l.Bytes()
	}

	max := 0
	t := ast.Token{Tok: SPACE, Lit: ""} // will be skipped when nothing matches, happens on newlines or empty lines
	for _, s := range syms {
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
				t = ast.Token{Tok: s.tok, Lit: string(lit)}
			}
		}
	}

	if max == len(l.buf) {
		t.Newline = true
	}
	if max < len(l.buf)-1 && l.buf[max+1] == '\n' {
		t.Newline = true
	}
	l.col += max
	l.buf = l.buf[max:]
	return t
}

// copied from std lib
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, trim(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), trim(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func trim(data []byte) []byte {
	if len(data) > 0 {
		return bytes.TrimSpace(data)
	}
	return data
}
