package parse

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

func Body(b *rd.Builder) (ok bool) {
	b.Enter("Body")
	defer b.Exit(&ok)

	if !Match(b, token.T{Type: chroma.Keyword, Value: "body"}) {
		return false
	}
	if !MatchType(b, chroma.Keyword) {
		return false
	}
	if !MatchType(b, chroma.NameFunction) && !MatchType(b, chroma.Keyword) {
		return false
	}
	// if next is ( -> params, if { open the body
	if Peek(b, token.T{Type: chroma.Punctuation, Value: "("}) {
		if !ArgList(b) {
			return false
		}
	}

	Comments(b)

	// now we should see {
	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "{"}) {
		return false
	}

	defer Comments(b)
	return BodyBody(b) && Match(b, token.T{Type: chroma.Punctuation, Value: "}"})
}

func BodyBody(b *rd.Builder) (ok bool) {
	b.Enter("BodyBody")
	defer b.Exit(&ok)

	Comments(b)
	Macro(b)
More:
	// classguardselection and selections or just selections
	ClassGuardSelections(b)
	BodySelections(b)

	if !Peek(b, token.T{Type: chroma.Punctuation, Value: "}"}) {
		goto More
	}
	return true
}

func ClassGuardSelections(b *rd.Builder) (ok bool) {
	b.Enter("ClassGuardSelections")
	defer b.Exit(&ok)

	if !MatchType(b, chroma.NameClass) {
		return false
	}
	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "::"}) {
		return false
	}
	return BodySelections(b)
}

func BodySelections(b *rd.Builder) (ok bool) {
	b.Enter("BodySelections")
	defer b.Exit(&ok)

	for {
		Comments(b) // comments in between selections and trailing ones
		if !Selection(b) {
			return true
		}
		Macro(b)
	}
}

func Selection(b *rd.Builder) (ok bool) {
	b.Enter("Selection")
	defer b.Exit(&ok)

	Comments(b)
	if !MatchType(b, chroma.KeywordReserved) && !MatchType(b, chroma.KeywordType) {
		return false
	}

	if !FatArrow(b) {
		return false
	}
	return Rval(b) && b.Match(token.T{Type: chroma.Punctuation, Value: ";"})
}
