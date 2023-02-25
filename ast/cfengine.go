package ast

// Token is a lexigraphical token in CFEngine.
type Token struct {
	Tok     int      // Token identifier.
	Lit     string   // String parsed for this token.
	Comment []string // possible comment attached to Token, unformatted as such.
}

// CFEngine AST Types.
type (
	Specification struct{ Container }
	Bundle        struct{ Container }
	Body          struct{ Container }
	PromiseGuard  struct{ Container }
	Identifier    struct{ Leaf }
	ClassGuard    struct{ Container }
	Promiser      struct{ Container }
	Selection     struct{ Container }
	Qstring       struct{ Leaf }
	Constraint    struct{ Container }
	FatArrow      struct{ Leaf }
	ThinArrow     struct{ Leaf }
	Function      struct{ Container }
	GiveArgItem   struct{ Leaf }
	NakedVar      struct{ Leaf }
	List          struct{ Container }
	ListItem      struct{ Leaf }
	ArgList       struct{ Container }
	ArgListItem   struct{ Leaf }
)
