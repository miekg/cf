package token

// Extra tokens we use in CFEngine, these tokens are slightly higher level than the tokens Chroma
// returns. For instance LiteralStrings are group together to become Qstrings.
const (
	None = iota - 1000
	Bundle
	Body
	Promise
	FatArrow
	ThinArrow
	Qstring
	NakedVar
	Comment // not in cfparse.y
	Identifier
)
