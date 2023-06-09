package parse

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/miekg/cf/internal/rd"
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
	const expect = `token.T {Keyword bundle}
token.T {Keyword agent}
token.T {NameFunction gitlab_server}
token.T {Punctuation {}
token.T {NameClass IsMattermostServer}
token.T {Punctuation ::}
token.T {TokenType(-994) "/var/opt/gitlab/mattermost/config.json"}
token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) "mattermost see https://cncz.ru.nl/#more/procedures/GitLab/#mattermost-configuratie"}
token.T {Punctuation ,}
token.T {KeywordReserved perms}
token.T {TokenType(-996) =>}
token.T {NameFunction mog}
token.T {Punctuation (}
token.T {LiteralNumberInteger 0660}
token.T {Punctuation ,}
token.T {NameFunction mattermost}
token.T {Punctuation ,}
token.T {NameFunction root}
token.T {Punctuation )}
token.T {Punctuation ;}
token.T {Punctuation }}
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
	const expect = `token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) 'mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie'}
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
	const expect = `token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) 'mattermost see https://cncz.ru.nl/more/procedures/#GitLab/mattermost-configuratie'}
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
	const expect = `token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) ` + backtick + `mattermost see https://cncz.ru.nl/more/procedures/GitLab/mattermost-configuratie` + backtick + "}\n"

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
	const expect = `token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) ` + backtick + `mattermost see https://cncz.ru.nl/more/procedures/GitLab/#mattermost-configuratie` + backtick + "}\n"

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
	const expect = `token.T {KeywordReserved inform}
token.T {TokenType(-996) =>}
token.T {NameVariable $(compounds.to_inform)}
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

func TestLexSingleQuotePunctuation(t *testing.T) {
	const input = `"lines" slist => { '#controlled by cfengine',
				};`
	const expect = `token.T {TokenType(-994) "lines"}
token.T {KeywordReserved slist}
token.T {TokenType(-996) =>}
token.T {Punctuation {}
token.T {TokenType(-994) '#controlled by cfengine'}
token.T {Punctuation ,}
token.T {Punctuation }}
token.T {Punctuation ;}
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

func TestLexSingleQuoteMultipleWords(t *testing.T) {
	const input = `comment => 'Ensure that the given parameter for file "$(file)" has only
the contents of the given parameter for content "$(content)"';`
	const expect = `token.T {KeywordReserved comment}
token.T {TokenType(-996) =>}
token.T {TokenType(-994) 'Ensure that the given parameter for file "$(file)" has only
the contents of the given parameter for content "$(content)"'}
token.T {Punctuation ;}
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
