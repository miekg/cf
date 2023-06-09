package parse

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

func Equal(a rd.Token, t token.T) bool {
	return a.(token.T).Type == t.Type && a.(token.T).Value == t.Value
}
func EqualType(a rd.Token, tt chroma.TokenType) bool { return a.(token.T).Type == tt }

// MatchType matches the next token type to tt. If OK the token is added to the tree.
func MatchType(b *rd.Builder, tt chroma.TokenType) bool {
	next, ok := b.Peek(1)
	if !ok {
		return false
	}
	if EqualType(next, tt) {
		b.Next()
		b.Add(next)
		return true
	}
	b.ErrorToken = next
	return false
}

// Match matches the next token type to tt. If OK the token is added to the tree.
func Match(b *rd.Builder, t token.T) bool {
	next, ok := b.Peek(1)
	if !ok {
		return false
	}
	if Equal(next, t) {
		b.Next()
		b.Add(next)
		return true
	}
	b.ErrorToken = next
	return false
}

// MatchDiscard matches the next token type to tt. If OK the token is not added to the tree, but discarded.
func MatchDiscard(b *rd.Builder, t token.T) bool {
	next, ok := b.Peek(1)
	if !ok {
		return false
	}
	if Equal(next, t) {
		b.Next()
		return true
	}

	b.ErrorToken = next
	return false
}

// Peek will peek the next token.
func Peek(b *rd.Builder, t token.T) bool {
	next, ok := b.Peek(1)
	if !ok {
		return false
	}
	return Equal(next, t)
}

// PeekClassGuard checks for a classguard by peeking the next 2 tokens.
func PeekClassGuard(b *rd.Builder) (ok bool) {
	peek1, ok := b.Peek(1)
	if !ok {
		return false
	}
	if !EqualType(peek1, chroma.NameClass) {
		return false
	}
	peek2, ok := b.Peek(2)
	if !ok {
		return false
	}
	if !Equal(peek2, token.T{Type: chroma.Punctuation, Value: "::"}) {
		return false
	}
	return true
}

// PeekPromiseGuard checks for a classguard by peeking the next 2 tokens.
func PeekPromiseGuard(b *rd.Builder) (ok bool) {
	peek1, ok := b.Peek(1)
	if !ok {
		return false
	}
	if !EqualType(peek1, chroma.KeywordDeclaration) {
		return false
	}
	peek2, ok := b.Peek(2)
	if !ok {
		return false
	}
	if !Equal(peek2, token.T{Type: chroma.Punctuation, Value: ":"}) {
		return false
	}
	return true
}
