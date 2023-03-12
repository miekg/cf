package parse

import (
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
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

	// Chroma has a bug where it sees a comment in a string and makes it an actual Comment, instead of properly
	// saying it is part of the string. This happens for single quote, and probably also backtick.
	// Compresses LiteralString* into a single Qstring, same for Comments. And fix backtick and quotes.
	pt := chroma.Token{Type: token.None}
	q := chroma.Token{Type: token.None} // qstring for single quotes and backtick string
	//defer println("*****")
	for _, t := range iter.Tokens() {
		//fmt.Printf("%T %v\n", t, t)
		switch t.Type {
		case chroma.LiteralString, chroma.LiteralStringInterpol, chroma.LiteralStringEscape:
			if q.Type == token.Qstring { // in a qstring gathering phase
				q.Value += t.Value
				continue
			}

			if pt.Type != token.Qstring && pt.Type != token.None {
				tokens = append(tokens, rd.Token(pt))
				pt.Value = ""
			}
			pt.Type = token.Qstring
			pt.Value += t.Value

		case chroma.Comment:
			if q.Type == token.Qstring { // we're in a single or backtick string, this comment closes both
				q.Value += t.Value
				// We match too match, so strip back to the closing quote.
				// Check first char for quote, indexLast with that value.
				begin := string(q.Value[0])
				if end := strings.LastIndex(q.Value, begin); end > 0 {
					q.Value = q.Value[:end+1]
				}

				q.Value = strings.TrimSuffix(q.Value, "\n") // TODO(miek): unix only now
				tokens = append(tokens, q)
				q.Type = token.None
				q.Value = ""
				continue
			}

			if pt.Type != token.Comment && pt.Type != token.None {
				tokens = append(tokens, rd.Token(pt))
				pt.Value = ""
			}

			pt.Type = token.Comment
			pt.Value += t.Value

		case chroma.Error:
			// quote
			if t.Value == "'" && q.Type == token.None { // open
				q.Type = token.Qstring
				q.Value = "'"
				continue
			}
			if t.Value == "'" && q.Type == token.Qstring { // close
				q.Value += "'"
				tokens = append(tokens, q)
				q.Type = token.None
				q.Value = ""
				continue
			}
			// backtick
			if t.Value == "`" && q.Type == token.None { // open
				q.Type = token.Qstring
				q.Value = "`"
				continue
			}
			if t.Value == "`" && q.Type == token.Qstring { // close
				q.Value += "`"
				tokens = append(tokens, q)
				q.Type = token.None
				q.Value = ""
				continue
			}

			if q.Type == token.Qstring { // append
				q.Value += t.Value
			}

			if q.Type == token.None {
				tokens = append(tokens, rd.Token(t))
			}

		case chroma.Operator:
			if t.Value == "=>" {
				tokens = append(tokens, rd.Token(chroma.Token{Type: token.FatArrow, Value: t.Value}))
			}

		case chroma.Text:
			if q.Type == token.Qstring {
				q.Value += t.Value
				continue
			}

			if pt.Type != token.None {
				tokens = append(tokens, pt)
			}
			pt.Type = token.None
			pt.Value = ""

		default:
			if q.Type == token.Qstring {
				q.Value += t.Value
				continue
			}

			if pt.Type != token.None {
				tokens = append(tokens, pt)
			}
			pt.Type = token.None
			pt.Value = ""

			tokens = append(tokens, rd.Token(t))
		}
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

	var tokens3 []rd.Token
	{
		// And another, which should also be a Qstring
		// chroma.Token {Punctuation "}
		// chroma.Token {NameVariable installed_names_canonified}
		// chroma.Token {Punctuation "}
		suppress := 0
		for i, t := range tokens2 {
			if t.(chroma.Token).Type == chroma.Punctuation && t.(chroma.Token).Value == `"` {
				if i < len(tokens2)-2 {
					if tokens2[i+1].(chroma.Token).Type == chroma.NameVariable {
						if tokens2[i+2].(chroma.Token).Type == chroma.Punctuation && tokens2[i+2].(chroma.Token).Value == `"` {
							// make this a indentifier, because arglist.cf uses it in that place.
							// there might be other places as well.
							suppress = 2
							tokens3 = append(tokens3, chroma.Token{Type: token.Qstring, Value: `"` + tokens2[i+1].(chroma.Token).Value + `"`})
							continue
						}
					}
				}
			}
			if suppress == 0 {
				tokens3 = append(tokens3, t)
			}
			if suppress > 0 {
				suppress--
			}

		}
	}

	return tokens3, nil
}
