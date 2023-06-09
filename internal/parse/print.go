package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

const _Space = "  "

// Printer used for some housekeeping during printing.
type Printer struct {
	first bool
	body  bool

	multilineList bool
}

// Print pretty prints the CFengine AST in tree.
func Print(w io.Writer, tree *rd.Tree) {
	if tree == nil {
		return // empty spec
	}

	p := &Printer{}
	remove(tree)
	align(tree)

	tw := &tw{w: &bytes.Buffer{}, width: 120} // make option?
	for _, t := range tree.Subtrees {
		p.print(tw, t, 0, tree)
	}
	// cleanup trailing whitespace while copying to writer
	scanner := bufio.NewScanner(tw.w)
	for scanner.Scan() {
		// Get Bytes and display the byte.
		b := scanner.Bytes()
		w.Write(bytes.TrimRight(b, " "))
		io.WriteString(w, "\n")
	}
}

func (p *Printer) print(w *tw, t *rd.Tree, depth int, parent *rd.Tree) {
	indent := ""
	if depth >= 1 {
		indent = strings.Repeat(_Space, depth-1)
	}

	// On Enter
	switch v := t.Data().(type) {
	case string:
		switch v {
		case "BundleBody":
			fmt.Fprintf(w, "\n{\n") // open the bundle
			p.body = false

		case "BodyBody":
			fmt.Fprintf(w, "\n{\n") // open the body
			p.body = true

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
					if prevOfType(parent, t, "Comment") { // we have insert a new line then
						fmt.Fprintf(w, "%s", indent)
					} else {
						fmt.Fprint(w, " ")
					}
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
			litems := countOfType(t, "Litem")
			comments := countOfType(t, "Comment")
			if litems+comments <= 10 && comments > 0 {
				p.multilineList = true
			}

		case "Litem":
		}

	case token.T:
		switch v.Type {
		case chroma.Keyword:
			switch v.Value {
			case "bundle", "body":
				if p.first {
					fmt.Fprintln(w)
				}
				p.first = true
				fmt.Fprintf(w, "%s ", v.Value)
			default:
				fmt.Fprintf(w, "%s ", v.Value)
			}

		case chroma.CommentPreproc:
			fmt.Fprintf(w, "\n%s", v.Value)

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

		case chroma.Comment:
			// Comments are nested as a child of ClassPromise. This makes them slighty too indented by one
			// step. Fix that here. FIX(miek).
			switch depth {
			case 1:
				if p.first { // top-level comments
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "%s", v.Value) // no indentation
			case 2: // comments between bundle and opening {
				if p.first { // top-level comments
					fmt.Fprintln(w)
				}
				if w.col > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintf(w, "%s", v.Value) // no indentation
				// small bug where this as a new line before the opening brace
			default:
				if w.col > 0 { // we've already outputted a line, this comment comes after the text, indent by _Space
					fmt.Fprintf(w, "%s%s", _Space, indentMultilineComment(v.Value, _Space))
				} else {
					cindent := indent[:len(indent)-2]
					if p.body && len(cindent) >= 2 {
						cindent = cindent[:len(cindent)-2]
					}
					fmt.Fprintf(w, "%s%s", cindent, indentMultilineComment(v.Value, cindent))
				}
				// comment in listem
				if w.bracecol > -1 {
					lindent := strings.Repeat(" ", w.bracecol)
					fmt.Fprintf(w, "%s", lindent)
				}
			}
			if strings.HasPrefix(v.Value, "# cffmt:list-nl") {
				p.multilineList = true
			}

		case token.Qstring:
			// TODO(miek): Needs indenting if spread over multiple lines. Possibly we need to strip prefix
			// whitespace.
			// Not added for now
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
			// there are constraints following this promisee, add a newline
			// if not don't, so the closing ; will be put on the same line.
			if len(parent.Subtrees) > 2 {
				fmt.Fprintf(w, "\n%s", indent)
			}

		case "ClassGuardPromises":

		case "ClassGuardSelections":

		case "Promise":
			last := lastOfType(parent, t, "Promise")
			single := promisersAllHaveSingleConstraint(parent)
			fmt.Fprint(w, ";\n")
			if !last && !single {
				fmt.Fprintf(w, "\n")
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
			p.multilineList = false

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
			p.multilineList = false

		case "Litem":
			last := lastOfType(parent, t, "Litem")
			if !last {
				fmt.Fprintf(w, ", ")
				if p.multilineList {
					lindent := strings.Repeat(" ", w.bracecol)
					fmt.Fprintf(w, "\n%s", lindent)
				}
			}

			if w.col > w.width && !last { // !last to prevent lonely '}' on the line
				lindent := strings.Repeat(" ", w.bracecol)
				fmt.Fprintf(w, "\n%s", lindent)
			}
		}

	case token.T:
		switch v.Type {
		case chroma.Keyword:

		case chroma.Punctuation:

		default:
		}

	default:
		panic("should not happen")
	}
}
