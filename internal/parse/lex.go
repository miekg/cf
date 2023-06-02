package parse

import (
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
	pt := chroma.Token{Type: token.None}
	//defer println("*****")
	for _, t := range iter.Tokens() {
		//fmt.Printf("%T %v\n", t, t)
		switch t.Type {
		case chroma.LiteralString, chroma.LiteralStringInterpol, chroma.LiteralStringEscape:
			if pt.Type != token.Qstring && pt.Type != token.None {
				tokens = append(tokens, rd.Token(pt))
				pt.Value = ""
			}
			pt.Type = token.Qstring
			pt.Value += t.Value

		case chroma.Operator:
			if t.Value == "=>" {
				tokens = append(tokens, rd.Token(chroma.Token{Type: token.FatArrow, Value: t.Value}))
			}
			if t.Value == "->" {
				tokens = append(tokens, rd.Token(chroma.Token{Type: token.ThinArrow, Value: t.Value}))
			}

		case chroma.Text:
			if pt.Type != token.None {
				tokens = append(tokens, pt)
			}
			pt.Type = token.None
			pt.Value = ""

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
