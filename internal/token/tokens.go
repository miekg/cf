package token

// Extra tokens we use in CFEngine, these tokens are slightly higher level than the tokens Chroma
// returns. For instance LiteralStrings are group together to become Qstrings.
const (
	None      = iota - 1000
	FatArrow  = -996
	ThinArrow = -995
	Qstring   = -994
)
