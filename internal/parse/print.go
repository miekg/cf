package parse

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/miekg/cf/token"
	"github.com/shivamMg/rd"
)

const _Space = "  "

type Printer struct {
	nostart bool // we we not outputting the first body/bundle of the file
	body    bool // are we in a body or bundle (indendation hack for comments)
}

// Print pretty prints the CFengine AST in tree.
func Print(w io.Writer, tree *rd.Tree) {
	if tree == nil {
		return // empty spec
	}

	align(tree)

	p := &Printer{}

	tw := &tw{w: w, width: 120} // make option?
	for _, t := range tree.Subtrees {
		p.print(tw, t, 0, tree)
	}
}

func (p Printer) print(w *tw, t *rd.Tree, depth int, parent *rd.Tree) {
	indent := ""
	if depth >= 0 {
		indent = strings.Repeat(_Space, depth)
	}

	// On Enter
	switch v := t.Data().(type) {
	case string:
		switch v {
		case "Bundle":
			p.body = false

		case "Body":
			p.body = true

		case "BundleBody", "BodyBody":
			fmt.Fprintf(w, "\n{\n") // open the bundle

		case "PromiseGuard":
			first := firstOfType(parent, t, "PromiseGuard")
			if !first {
				fmt.Fprintf(w, "\n\n")
			}
			// Maybe we've already printed classguards, in that case we also want 2 newlines.
			if sequenceOfChild(parent, t) > 0 && first {
				fmt.Fprintf(w, "\n\n")
			}
			fmt.Fprintf(w, "%s", indent)

		case "ClassGuardPromises":
			first := firstOfType(parent, t, "ClassGuardPromises")
			if !first {
				fmt.Fprintln(w)
			}
			printChildrenOfType(w, t, chroma.NameClass, func(v string) {
				fmt.Fprintf(w, "%s%s::\n", indent, v)
			})

		case "ClassGuardSelections":
			seq := sequenceOfChild(parent, t)
			if seq != 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w)
			}
			printChildrenOfType(w, t, chroma.NameClass, func(v string) {
				fmt.Fprintf(w, "%s%s::\n", indent, v)
			})

		case "Promise":
			fmt.Fprintf(w, "%s", indent)
			printFirstChildOfType(w, t, token.Qstring, func(v string) {
				v1 := indentMultilineQstring(v, indent)
				fmt.Fprintf(w, "%s", v1)
			})

		case "Selection":
			// We indent too much because BodySelection is in the AST
			// remove 2 spaces from indent.
			fmt.Fprintf(w, "%s", indent[:len(indent)-2])

		case "Constraint":
			single := countOfType(parent, "Constraint") == 1
			if single {
				if constraintPreventSingleLine(t) {
					fmt.Fprintf(w, "\n%s", indent)
				} else {
					fmt.Fprint(w, " ")
				}
			} else {
				fmt.Fprintf(w, "\n%s", indent)
			}

		case "ArgList":
			fmt.Fprintf(w, "(")

		case "Aitem":

		case "GiveArgList":
			fmt.Fprintf(w, "(")

		case "GaItem":

		case "List":
			if len(t.Subtrees) == 0 {
				fmt.Fprintf(w, "{")
			} else {
				fmt.Fprintf(w, "{ ")
			}
			w.bracecol = w.col

		case "Litem":
		}

	case chroma.Token:
		switch v.Type {
		case chroma.Keyword:
			switch v.Value {
			case "bundle", "body":
				if p.nostart {
					fmt.Fprintln(w)
				}
				p.nostart = true
				fmt.Fprintf(w, "%s ", v.Value)
			default:
				fmt.Fprintf(w, "%s ", v.Value)
			}
		case chroma.KeywordDeclaration:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.KeywordReserved:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.KeywordType:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.NameFunction:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.NameVariable:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.NameClass:
			fmt.Fprintf(w, "%s", v.Value)

		case chroma.LiteralNumberInteger:
			fmt.Fprintf(w, "%s", v.Value)

		case token.Comment:
			// Comments are nested as a child of ClassPromise. This makes them slighty too indented by one
			// step. Fix that here. FIX(miek).
			switch depth {
			case 1:
				if p.nostart { // top-level comments
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "%s", v.Value) // no indentation
			case 2: // comments between bundle and opening {
				if p.nostart { // top-level comments
					fmt.Fprintln(w)
				}
				if w.col > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "%s", v.Value) // no indentation
			default:
				if w.col > 0 { // we've already outputted a line, this comment comes after the text, indent by _Space
					fmt.Fprintf(w, "%s%s", _Space, eolComment(v.Value))
				} else {
					// todo hack, in a body, we indent too much here, need -4 instead of -2
					cindent := depth - 1
					if p.body {
						cindent = depth - 2
					}
					fmt.Fprintf(w, "%s", indentMultilineComment(v.Value, cindent))
				}
				// comment in listem
				if w.bracecol > -1 {
					lindent := strings.Repeat(" ", w.bracecol)
					fmt.Fprintf(w, "%s", lindent)
				}
			}

		case token.Qstring:
			// TODO(miek): Needs indenting ever??
			// Not added for now
			fmt.Fprintf(w, "%s", v.Value)

		case token.NakedVar:
			fmt.Fprintf(w, "%s", v.Value)

		case token.FatArrow:
			fmt.Fprintf(w, " %s ", v.Value)

		case token.ThinArrow:
			fmt.Fprintf(w, " %s ", v.Value)

		case chroma.Punctuation:

		default:
			fmt.Fprintf(w, "%v\n", v)
		}

	default:
		panic("should not happen")
	}

	for _, c := range t.Subtrees {
		p.print(w, c, depth+1, t)
	}

	// On Leave
	switch v := t.Data().(type) {
	case string:
		switch v {
		case "BundleBody":
			fmt.Fprintf(w, "}\n")

		case "BodyBody":
			fmt.Fprintf(w, "\n}\n") // needs extra new line

		case "PromiseGuard":
			fmt.Fprint(w, ":\n\n")

		case "Promisee":
			fmt.Fprintf(w, "\n%s", indent)

		case "ClassGuardPromises":

		case "ClassGuardSelections":

		case "Promise":
			last := lastOfType(parent, t, "Promise")
			fmt.Fprint(w, ";\n")
			if !last {
				fmt.Fprintln(w)
			}

		case "Selection":
			last := lastOfType(parent, t, "Selection")
			fmt.Fprint(w, ";")
			if !last {
				fmt.Fprintln(w)
			}

		case "Constraint":
			last := lastOfType(parent, t, "Constraint")
			if !last {
				fmt.Fprint(w, ",")
			}

		case "ArgList":
			fmt.Fprintf(w, ")")

		case "Aitem":
			last := lastOfType(parent, t, "Aitem")
			if !last {
				fmt.Fprintf(w, ",")
			}

		case "GiveArgList":
			fmt.Fprintf(w, ")")

		case "GaItem":
			last := lastOfType(parent, t, "GaItem")
			if !last {
				fmt.Fprintf(w, ", ")
			}

		case "List":
			if len(t.Subtrees) == 0 {
				fmt.Fprintf(w, "}")
			} else {
				fmt.Fprint(w, " }")
			}
			w.bracecol = -1

		case "Litem":
			last := lastOfType(parent, t, "Litem")
			if !last {
				fmt.Fprintf(w, ", ")
			}
			if w.col > w.width {
				lindent := strings.Repeat(" ", w.bracecol)
				fmt.Fprintf(w, "\n%s", lindent)
			}
		}

	case chroma.Token:
		switch v.Type {
		case chroma.Keyword:

		case chroma.Punctuation:

		default:
		}

	default:
		panic("should not happen")
	}
}
