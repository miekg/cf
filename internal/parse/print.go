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

var first bool

// Print pretty prints the CFengine AST in tree.
func Print(w io.Writer, tree *rd.Tree) {
	if tree == nil {
		return // empty spec
	}

	align(tree)

	tw := &tw{w: w}
	for _, t := range tree.Subtrees {
		print(tw, t, 0, tree)
	}
}

func print(w *tw, t *rd.Tree, depth int, parent *rd.Tree) {
	indent := ""
	if depth >= 0 {
		indent = strings.Repeat(_Space, depth)
	}

	// On Enter
	switch v := t.Data().(type) {
	case string:
		switch v {
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

		case "Promise":
			fmt.Fprintf(w, "%s", indent)
			printFirstChildOfType(w, t, token.Qstring, func(v string) {
				v1 := indentMultilineQstring(v, indent)
				fmt.Fprintf(w, "%s", v1)
			})

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
			fmt.Fprintf(w, "{ ")

		case "Litem":
		}

	case chroma.Token:
		switch v.Type {
		case chroma.Keyword:
			switch v.Value {
			case "bundle", "body":
				if first {
					fmt.Fprintln(w)
				}
				first = true
				fmt.Fprintf(w, "%s", v.Value)
			default:
				fmt.Fprintf(w, " %s", v.Value)
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
			// Comments are nested as a child of  ClassPromise. This makes them slighty too indented by one
			// step. Fix that here. FIX(miek).
			if w.col > 0 { // we've already outputted a line, this comment comes after the text, indent by _Space
				fmt.Fprintf(w, "%s%s", _Space, v.Value)
			} else {
				cindent := indent[:len(indent)-2]
				fmt.Fprintf(w, "%s%s", cindent, v.Value)
			}

		case token.Qstring:
			// TODO(miek): Needs indenting ever??
			// Not added for now
			fmt.Fprintf(w, "%s", v.Value)

		case token.NakedVar:
			fmt.Fprintf(w, "%s", v.Value)

		case token.FatArrow:
			fmt.Fprintf(w, " %s ", v.Value)

		case chroma.Punctuation:

		default:
			fmt.Fprintf(w, "%v\n", v)
		}

	default:
		panic("should not happen")
	}

	for _, c := range t.Subtrees {
		print(w, c, depth+1, t)
	}

	// On Leave
	switch v := t.Data().(type) {
	case string:
		switch v {
		case "BundleBody", "BodyBody":
			fmt.Fprintf(w, " }\n")

		case "PromiseGuard":
			fmt.Fprint(w, ":\n\n")

		case "ClassGuardPromises":

		case "Promise":
			last := lastOfType(parent, t, "Promise")
			fmt.Fprint(w, ";\n")
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
			fmt.Fprintf(w, " }")

		case "Litem":
			last := lastOfType(parent, t, "Litem")
			if !last {
				fmt.Fprintf(w, ", ")
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
