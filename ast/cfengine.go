package ast

// Token is a lexigraphical token in CFEngine.
type Token struct {
	Tok     int      // Token identifier.
	Lit     string   // String parsed for this token.
	Comment []string // Possible comment attached to Token, unformatted as such.
	Newline bool     // We have a newline at the end, important for multiline qstrings.
}

// CFEngine AST Types.
type (
	Specification struct{ Container }
	Bundle        struct{ Container }
	Body          struct{ Container }
	PromiseGuard  struct{ Container }
	ClassGuard    struct{ Container }
	Promiser      struct{ Container }
	Selection     struct{ Container }
	Qstring       struct{ Leaf }
	FatArrow      struct{ Leaf }
	ThinArrow     struct{ Leaf }
	NakedVar      struct{ Leaf }
	Identifier    struct{ Leaf }
	Function      struct{ Container }
	GiveArgItem   struct{ Container }
	List          struct{ Container }
	ListItem      struct{ Container }
	ArgList       struct{ Container }
	ArgListItem   struct{ Container }
	Constraint    struct {
		Container
		SingleLine bool // True when we want to print this on a single line.
	}
)
