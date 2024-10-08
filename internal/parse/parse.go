package parse

import (
	"fmt"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

// We have an infinite parsing loop somewhere, if we break it we get parsing
// errors in the test data we have....Meaning we depend on a 'return true' that
// should be a 'false'
// Just quit after 1000(?) loops, until we get a good handle on this. Stuff is fast
// enough, that you don't notice and you get a nice error.
const LoopBreak = 1000

func Specification(b *rd.Builder) (ok bool) {
	b.Enter("Specification")
	defer b.Exit(&ok)

	Comments(b)

More:
	ok1 := Bundle(b)
	ok2 := Body(b)
	if !ok1 && !ok2 {
		return false
	}

	if _, ok := b.Peek(1); !ok {
		return ok1 || ok2
	}
	Comments(b)
	goto More
}

func Bundle(b *rd.Builder) (ok bool) {
	b.Enter("Bundle")
	defer b.Exit(&ok)

	if !Match(b, token.T{Type: chroma.Keyword, Value: "bundle"}) {
		return false
	}
	if !MatchType(b, chroma.Keyword) {
		return false
	}
	if !MatchType(b, chroma.NameFunction) {
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
	return BundleBody(b) && MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "}"})
}

func BundleBody(b *rd.Builder) (ok bool) {
	b.Enter("BundleBody")
	defer b.Exit(&ok)

	Comments(b)
	Macro(b)
	i := 0
More:
	i++
	// Zero or more promiseguards (single : ) and then zero more classpromises.
	ok1 := PromiseGuard(b)
	ok2 := ClassPromises(b)
	if !ok1 && !ok2 {
		return false
	}

	if PeekPromiseGuard(b) {
		if i > LoopBreak {
			return false
		}
		goto More
	}
	// PeekClassPromise as well TODO(miek)

	return true
}

func ClassPromises(b *rd.Builder) (ok bool) {
	b.Enter("ClassPromises")
	defer b.Exit(&ok)

	// various possiblities:
	//
	// classguard::
	// promises
	// or
	// < no classguard>
	// promises

More:
	ClassGuardPromises(b)
	Promises(b)
	if PeekClassGuard(b) {
		goto More
	}
	return true
}

func ClassGuardPromises(b *rd.Builder) (ok bool) {
	b.Enter("ClassGuardPromises")
	defer b.Exit(&ok)

	if !MatchType(b, chroma.NameClass) {
		return false
	}
	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "::"}) {
		return false
	}
	return Promises(b)
}

func Promises(b *rd.Builder) (ok bool) {
	// b.Enter("Promises") - not in AST
	// b.Exit(&ok)

	for {
		Comments(b) // comments in between promises and trailing ones
		if !Promise(b) {
			return true
		}
		Macro(b)
	}
}

func Promise(b *rd.Builder) (ok bool) {
	b.Enter(token.Promise)
	defer b.Exit(&ok)

	if !MatchType(b, token.Qstring) {
		return false
	}
	Promisee(b)
	Comments(b)

	// zero or more constraints, and then a closing ;
	return PromiseConstraints(b) && Match(b, token.T{Type: chroma.Punctuation, Value: ";"})
}

func PromiseConstraints(b *rd.Builder) (ok bool) {
	b.Enter("PromiseConstraints")
	b.Exit(&ok)

	// if no constraint found, we don't have any, this is OK
	ok = Peek(b, token.T{Type: chroma.Punctuation, Value: ";"})
	if ok {
		return true // empty contraints list
	}
More:
	Constraint(b)
	Comments(b)
	// next token is , we have more Constraints, otherwise return
	if ok = Match(b, token.T{Type: chroma.Punctuation, Value: ","}); ok {
		Comments(b)
		goto More
	}
	return true

}

func Constraint(b *rd.Builder) (ok bool) {
	b.Enter(token.Constraint)
	defer b.Exit(&ok)

	Comments(b)
	if !MatchType(b, chroma.KeywordReserved) && !MatchType(b, chroma.KeywordType) {
		return false
	}

	if !FatArrow(b) {
		return false
	}
	return Rval(b)
}

func Promisee(b *rd.Builder) (ok bool) {
	b.Enter("Promisee")
	defer b.Exit(&ok)

	if !ThinArrow(b) {
		return false
	}
	return Rval(b)
}

func Rval(b *rd.Builder) (ok bool) {
	b.Enter("Rval")
	defer b.Exit(&ok)
	if Qstring(b) {
		return true
	}
	if Function(b) {
		return true
	}
	if List(b) {
		return true
	}
	// Identifier, NameVariable here too?
	if MatchType(b, chroma.NameFunction) {
		return true
	}
	if MatchType(b, chroma.LiteralNumberInteger) {
		return true
	}
	if NakedVar(b) {
		return true
	}
	if NamespaceFunction(b) {
		return true
	}
	return false
}

// NamespaceFunction parses the (relative?) new syntax "default:function", this is lexed as:
// KeywordDeclaration Punctuation(:) NameFunction. This functions check for that.
// It adds a Function to the AST.
func NamespaceFunction(b *rd.Builder) (ok bool) {
	b.Enter("NamespaceFunction")
	defer b.Exit(&ok)
	if !MatchType(b, chroma.KeywordDeclaration) {
		return false
	}
	if !Match(b, token.T{Type: chroma.Punctuation, Value: ":"}) {
		return false
	}
	if !MatchType(b, chroma.NameFunction) {
		return false
	}
	return true
}

func Function(b *rd.Builder) (ok bool) {
	b.Enter("Function")
	defer b.Exit(&ok)
	if !MatchType(b, chroma.NameFunction) {
		return false
	}
	// Check for Identifier
	if !Peek(b, token.T{Type: chroma.Punctuation, Value: "("}) {
		return false // no ( after name, this is not a function
	}

	return GiveArgList(b)
}

func GiveArgList(b *rd.Builder) (ok bool) {
	b.Enter("GiveArgList")
	defer b.Exit(&ok)
	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "("}) {
		return false
	}
	return GaItems(b) && MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: ")"})
}

func GaItems(b *rd.Builder) (ok bool) {
	//b.Enter("GaItems")   - don't add in AST.
	//defer b.Exit(&ok)
	ok = Peek(b, token.T{Type: chroma.Punctuation, Value: ")"})
	if ok {
		return true // empty function arglist
	}

More:
	GaItem(b) // if !ok this is an actual error?

	// if next thing is a , we have another GaItems, otherwise return
	if ok = Match(b, token.T{Type: chroma.Punctuation, Value: ","}); ok {
		// TODO: should not add this to the AST.
		goto More
	}
	return true
}

func GaItem(b *rd.Builder) (ok bool) {
	b.Enter("GaItem")
	defer b.Exit(&ok)

	Comments(b)

	if Qstring(b) {
		return true
	}
	if Function(b) {
		return true
	}
	// Identifier
	// NameVariable here too?
	if MatchType(b, chroma.NameFunction) {
		return true
	}
	if MatchType(b, chroma.LiteralNumberInteger) {
		return true
	}
	if NakedVar(b) {
		return true
	}

	return false
}

func List(b *rd.Builder) (ok bool) {
	b.Enter(token.List)
	defer b.Exit(&ok)

	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "{"}) {
		return false
	}
	return Litems(b) && MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "}"})
}

func Litems(b *rd.Builder) (ok bool) {
	//b.Enter(token.Litems)   - don't add in AST.
	//defer b.Exit(&ok)
	ok = Peek(b, token.T{Type: chroma.Punctuation, Value: "}"})
	if ok {
		return true // empty list
	}

More:
	Comments(b)
	Litem(b)

	// next token is , we have more Litems, otherwise return
	if ok = Match(b, token.T{Type: chroma.Punctuation, Value: ","}); ok {
		goto More
	}
	return true
}

func Litem(b *rd.Builder) (ok bool) {
	b.Enter(token.Litem)
	defer b.Exit(&ok)

	// comments in lists work, with the current printing because then insert a new line
	// so it's at the end of the line.
	Comments(b)

	if Qstring(b) {
		return true
	}
	if Function(b) {
		return true
	}
	// Identifier
	if MatchType(b, chroma.NameFunction) {
		return true
	}
	if MatchType(b, chroma.LiteralNumberInteger) {
		return true
	}

	if NakedVar(b) {
		return true
	}

	return false
}

func ArgList(b *rd.Builder) (ok bool) {
	b.Enter("ArgList")
	defer b.Exit(&ok)

	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: "("}) {
		return false
	}
	return Aitems(b) && MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: ")"})
}

func Aitems(b *rd.Builder) (ok bool) {
	//b.Enter("Aitems")   - don't add in AST.
	//defer b.Exit(&ok)
	ok = Peek(b, token.T{Type: chroma.Punctuation, Value: ")"})
	if ok {
		return true // empty list.. fails currently
	}

More:
	Aitem(b)

	// next token is , we have more Aitems, otherwise return
	if ok = Match(b, token.T{Type: chroma.Punctuation, Value: ","}); ok {
		goto More
	}
	return true
}

func Aitem(b *rd.Builder) (ok bool) {
	b.Enter("Aitem")
	defer b.Exit(&ok)

	// Only Identifiers allowed.
	if ok = MatchType(b, chroma.NameVariable); ok {
		return true
	}
	if ok = MatchType(b, chroma.NameFunction); ok {
		return true
	}
	return false
}

func PromiseGuard(b *rd.Builder) (ok bool) {
	b.Enter("PromiseGuard")
	defer b.Exit(&ok)

	if !MatchType(b, chroma.KeywordDeclaration) {
		return false
	}
	if !MatchDiscard(b, token.T{Type: chroma.Punctuation, Value: ":"}) {
		return false
	}
	return true
}

func Qstring(b *rd.Builder) (ok bool) {
	b.Enter("Qstring")
	defer b.Exit(&ok)
	return MatchType(b, token.Qstring)
}

func NakedVar(b *rd.Builder) (ok bool) {
	b.Enter("NakedVar")
	defer b.Exit(&ok)

	return MatchType(b, chroma.NameVariable)
}

func FatArrow(b *rd.Builder) (ok bool) {
	b.Enter("FatArrow")
	defer b.Exit(&ok)

	return MatchType(b, token.FatArrow)
}

func ThinArrow(b *rd.Builder) (ok bool) {
	b.Enter("ThinArrow")
	defer b.Exit(&ok)

	return MatchType(b, token.ThinArrow)
}

func Comments(b *rd.Builder) (ok bool) {
	if !Comment(b) {
		return false
	}
	for {
		if !Comment(b) {
			return true
		}
	}
}

func Macro(b *rd.Builder) (ok bool) {
	b.Enter("Macro")
	defer b.Exit(&ok)

	return MatchType(b, chroma.CommentPreproc)
}

func Comment(b *rd.Builder) (ok bool) {
	b.Enter("Comment")
	defer b.Exit(&ok)

	return MatchType(b, chroma.Comment)
}

func Fmt(b *rd.Builder, a string, i int) {
	tok, ok := b.Peek(i)
	if !ok {
		return
	}
	fmt.Printf("%s %T %v\n", a, tok.(token.T), tok.(token.T))
}
