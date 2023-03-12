package parse

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/shivamMg/rd"
)

func TestLexDoubleQuote(t *testing.T) {
	const input = `bundle agent gitlab_server
{
  IsMattermostServer::
   "/var/opt/gitlab/mattermost/config.json"
		comment   => "mattermost see https://cncz.ru.nl/#more/procedures/GitLab/#mattermost-configuratie",
		perms     => mog(0660, mattermost, root);
}
`
	const expect = `chroma.Token {Keyword bundle}
chroma.Token {Keyword agent}
chroma.Token {NameFunction gitlab_server}
chroma.Token {Punctuation {}
chroma.Token {NameClass IsMattermostServer}
chroma.Token {Punctuation ::}
chroma.Token {TokenType(-994) "/var/opt/gitlab/mattermost/config.json"}
chroma.Token {KeywordReserved comment}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-994) "mattermost see https://cncz.ru.nl/#more/procedures/GitLab/#mattermost-configuratie"}
chroma.Token {Punctuation ,}
chroma.Token {KeywordReserved perms}
chroma.Token {TokenType(-996) =>}
chroma.Token {NameFunction mog}
chroma.Token {Punctuation (}
chroma.Token {LiteralNumberInteger 0660}
chroma.Token {Punctuation ,}
chroma.Token {NameFunction mattermost}
chroma.Token {Punctuation ,}
chroma.Token {NameFunction root}
chroma.Token {Punctuation )}
chroma.Token {Punctuation ;}
chroma.Token {Punctuation }}
`
	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func TestLexSingleQuote(t *testing.T) {
	// needs the newline
	const input = `comment   => 'mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie'
`
	const expect = `chroma.Token {KeywordReserved comment}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-994) 'mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie'}
`

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func TestLexSingleQuoteComment(t *testing.T) {
	// needs the newline
	const input = `comment   => 'mattermost see https://cncz.ru.nl/more/procedures/#GitLab/mattermost-configuratie'
`
	const expect = `chroma.Token {KeywordReserved comment}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-994) 'mattermost see https://cncz.ru.nl/more/procedures/#GitLab/mattermost-configuratie'}
`

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

const backtick = "`"

func TestLexBacktickQuote(t *testing.T) {
	// needs the newline
	const input = "comment   => " + backtick + "mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie" + backtick + "\n"
	const expect = `chroma.Token {KeywordReserved comment}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-994) ` + backtick + `mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie` + backtick + "}\n"

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func TestLexBacktickQuoteComment(t *testing.T) {
	// needs the newline
	const input = "comment   => " + backtick + "mattermost see https://cncz.ru.nl/more/procedures/GitLab/#mattermost-configuratie" + backtick + "\n"
	const expect = `chroma.Token {KeywordReserved comment}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-994) ` + backtick + `mattermost see https://cncz.ru.nl/more/procedures/GitLab/#mattermost-configuratie` + backtick + "}\n"

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func TestLexNakedVar(t *testing.T) {
	const input = "inform => $(compounds.to_inform)\n"
	const expect = `chroma.Token {KeywordReserved inform}
chroma.Token {TokenType(-996) =>}
chroma.Token {TokenType(-993) $(compounds.to_inform)}
`

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func TestLexSinglekQuotePunctuation(t *testing.T) {
	const input = `"lines" slist => { '#controlled by cfengine',
				};`
	const expect = `chroma.Token {TokenType(-994) "lines"}
chroma.Token {KeywordType slist}
chroma.Token {TokenType(-996) =>}
chroma.Token {Punctuation {}
chroma.Token {TokenType(-994) '#controlled by cfengine'}
chroma.Token {Punctuation }}
chroma.Token {Punctuation ;}
`

	tokens, err := Lex(string(input))
	if err != nil {
		t.Fatal(err)
	}
	got := tokenToString(tokens)

	if got != expect {
		t.Errorf("Expected\n%s\n,Got\n%s\n", expect, got)
	}
}

func tokenToString(tokens []rd.Token) string {
	b := &bytes.Buffer{}
	for _, t := range tokens {
		fmt.Fprintf(b, "%T %v\n", t, t)
	}
	return b.String()
}
