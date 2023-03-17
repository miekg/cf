package parse

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/miekg/cf/token"
	"github.com/shivamMg/rd"
)

func Lex(specification string) ([]rd.Token, error) {
	lexer := lexers.Get("CFEngine3")
	var tokens []rd.Token
	iter, err := lexer.Tokenise(nil, specification)
	if err != nil {
		return nil, err
	}
	// There are several passes here. Mostly to not have a very complex single loop. Surely this can be optimized,
	// OTOH it's only iterating over a few thousand tokens and creating a bunch of garbage memory.

	// Compresses LiteralString* into a single Qstring, same for Comments.
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

		case chroma.Comment:
			if pt.Type != token.Comment && pt.Type != token.None {
				tokens = append(tokens, rd.Token(pt))
				pt.Value = ""
			}

			pt.Type = token.Comment
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

	var tokens2 []rd.Token
	{
		// To not complicate the above loop more we do another loop over the tokens to extract naked vars
		// $(..), shows up as Error($)Punctuation((). Grab those into a NakedVar
		//
		// chroma.Token {Error $}
		// chroma.Token {Punctuation (}
		// chroma.Token {NameFunction sys}
		// chroma.Token {Error .}
		// chroma.Token {NameFunction policy_hub}
		// chroma.Token {Punctuation )}
		nakedvar := chroma.Token{Type: token.None}
		for i, t := range tokens {
			// open
			if t.(chroma.Token).Type == chroma.Error && t.(chroma.Token).Value == "$" {
				if i < len(tokens)-1 && tokens[i+1].(chroma.Token).Type == chroma.Punctuation && tokens[i+1].(chroma.Token).Value == "(" {
					nakedvar.Type = token.NakedVar
				}
			}
			if nakedvar.Type == token.None {
				tokens2 = append(tokens2, t)
				continue
			}

			// close
			if t.(chroma.Token).Type == chroma.Punctuation && t.(chroma.Token).Value == ")" {
				nakedvar.Value += ")"
				tokens2 = append(tokens2, nakedvar)
				nakedvar = chroma.Token{Type: token.None}
				continue
			}

			nakedvar.Value += t.(chroma.Token).Value
		}
	}

	return tokens2, nil
}
