package parse

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

func Lex(specification string) ([]rd.Token, error) {
	lexer := lexers.Get("CFEngine3")
	var tokens []rd.Token
	iter, err := lexer.Tokenise(nil, specification)
	if err != nil {
		return nil, err
	}
	// Compresses LiteralString* into a single Qstring
	pt := token.T{Type: token.None}
	line := 1
	//defer println("*****")
	for _, t1 := range iter.Tokens() {
		t := token.T{Type: t1.Type, Value: t1.Value, Line: line}
		//fmt.Printf("%T %v\n", t, t)
		switch t.Type {
		case chroma.LiteralString, chroma.LiteralStringInterpol, chroma.LiteralStringEscape:
			if pt.Type != token.Qstring && pt.Type != token.None {
				tokens = append(tokens, rd.Token(pt))
				pt.Value = ""
			}
			pt.Type = token.Qstring
			pt.Value += t.Value
			pt.Line = line

		case chroma.Operator:
			if t.Value == "=>" {
				tokens = append(tokens, rd.Token(token.T{Type: token.FatArrow, Value: t.Value, Line: line}))
			}
			if t.Value == "->" {
				tokens = append(tokens, rd.Token(token.T{Type: token.ThinArrow, Value: t.Value, Line: line}))
			}

		case chroma.Text:
			if pt.Type != token.None {
				tokens = append(tokens, pt)
			}
			pt.Type = token.None
			pt.Value = ""
			line += strings.Count(t.Value, "\n")

		default:
			if pt.Type != token.None {
				tokens = append(tokens, pt)
			}
			pt.Type = token.None
			pt.Value = ""

			tokens = append(tokens, rd.Token(t))
		}
	}
	if pt.Type != token.None {
		tokens = append(tokens, pt)
	}

	return tokens, nil
}
