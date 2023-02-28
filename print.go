package cf

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/miekg/cf/ast"
)

// Print pretty prints the CFengine AST in doc.
func Print(w io.Writer, doc ast.Node) {
	wr := &tw{w: w, width: 100}
	for i, c := range doc.Children() {
		printRecur(wr, c, -1, i == 0, i == len(doc.Children())-1) // -1 because Specification is the top-level (noop) container.
	}
}

// PrintWithWidth pretty prints the CFengine AST in doc, but allows setting a custom width.
func PrintWithWidth(w io.Writer, width uint, doc ast.Node) {
	wr := &tw{w: w, width: int(width)}
	for i, c := range doc.Children() {
		printRecur(wr, c, -1, i == 0, i == len(doc.Children())-1) // -1 because Specification is the top-level (noop) container.
	}
}

func printDefault(w io.Writer, indent string, typeName string, token ast.Token) {
	if len(token.Lit) > 0 {
		if len(token.Comment) > 0 {
			for i := range token.Comment {
				fmt.Fprintf(w, "%2d %s->%s\n", len(indent), indent, token.Comment[i])
			}
		}
		fmt.Fprintf(w, "%2d %s%s '%s'\n", len(indent), indent, typeName, token.Lit)
	} else {
		for i := range token.Comment {
			fmt.Fprintf(w, "%2d %s->%s\n", len(indent), indent, token.Comment[i])
		}
		fmt.Fprintf(w, "%2d %s%s\n", len(indent), indent, typeName)
	}
}

const _Space = "  "

func printRecur(w io.Writer, node ast.Node, depth int, first, last bool) {
	if node == nil {
		return
	}
	indent := ""
	if depth >= 0 {
		indent = strings.Repeat(_Space, depth)
	}

	// Comments
	commentNoNewline := "\n"
	if len(node.Token().Comment) > 0 && depth > 0 {
		fmt.Fprintln(w)
	}
	for i := range node.Token().Comment {
		fmt.Fprintf(w, "%s%s\n", indent, node.Token().Comment[i])
		commentNoNewline = ""
	}

	// On Enter
	// Some nodes can be multline - but this is a relative experimental addition, so not sure which ones are that.
	switch v := node.(type) {
	case *ast.Specification: // start of the tree, ignore

	case *ast.Bundle, *ast.Body:
		if !first {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s", v.Token().Lit)
		printChildrenOfType(w, v, " %s", "*ast.Identifier")
		// First child is either ArgList or not, if not, end the bundle/body, otherwise print the list and end the
		// bundle
		children := v.Children()
		if len(children) > 0 {
			if _, ok := children[0].(*ast.ArgList); !ok {
				fmt.Fprint(w, "\n{")
			}
		}

	case *ast.PromiseGuard:
		fmt.Fprintf(w, "%s%s%s\n", commentNoNewline, indent, v.Token().Lit)

	case *ast.ClassGuard:
		fmt.Fprintf(w, "%s%s%s\n", commentNoNewline, indent, v.Token().Lit)

	case *ast.Promiser:
		children := len(v.Children()) != 0
		newline := ""
		if children {
			newline = "\n"
		}
		// if my parent is directly a promise guard, insert newline.
		promisenewline := ""
		if _, ok := v.Parent().(*ast.PromiseGuard); ok {
			promisenewline = "\n"
		}
		// this can be multiline
		multiline := strings.Replace(v.Token().Lit, "\n", "\n"+indent, -1)
		fmt.Fprintf(w, "%s%s%s%s", promisenewline, indent, multiline, newline)

	case *ast.Constraint:
		fmt.Fprintf(w, "%s%s", indent, v.Token().Lit)

	case *ast.Selection:
		fmt.Fprintf(w, "%s%s%s", commentNoNewline, indent, v.Token().Lit)

	case *ast.Function:
		printChildrenOfType(w, v, "%s", "*ast.Identifier")
		fmt.Fprint(w, "(")

	case *ast.GiveArgItem:
		fmt.Fprintf(w, "%s", v.Token().Lit)
		if !last {
			fmt.Fprint(w, ", ")
		}

	case *ast.List:
		printChildrenOfType(w, v, "%s", "*ast.Identifier")
		fmt.Fprint(w, "{ ")

	case *ast.ListItem:
		fmt.Fprintf(w, "%s", v.Token().Lit)
		// wrapping
		if w.(*tw).col > w.(*tw).width {
			if !last {
				fmt.Fprint(w, ", ")
				fmt.Fprintf(w, "\n%s%s", indent, _Space) // slightly deeper indent
			}
			break
		}
		if !last {
			fmt.Fprint(w, ", ")
		}

	case *ast.ArgList:
		// We receive the arg list in reverse, becauwe the grammar has 'item, items' (all other lists have
		// 'items, item). So reverse all children here (there can only be ArgListItem's)
		ast.Reverse(v)
		fmt.Fprint(w, "(")

	case *ast.ArgListItem:
		fmt.Fprintf(w, "%s", v.Token().Lit)
		if !last {
			fmt.Fprint(w, ", ")
		}

	case *ast.FatArrow, *ast.ThinArrow:
		fmt.Fprintf(w, " %s ", v.Token().Lit)

	case *ast.Qstring:
		// this can be multiline
		multiline := strings.Replace(v.Token().Lit, "\n", "\n"+indent, -1)
		fmt.Fprintf(w, "%s", multiline)

	case *ast.Identifier:
		fmt.Fprintf(w, "%s", v.Token().Lit)

	default:
		printDefault(w, indent, fmt.Sprintf("%T", v), v.Token())
	}

	for i, child := range node.Children() {
		printRecur(w, child, depth+1, i == 0, i == len(node.Children())-1)
	}

	// On Leave
	switch /*v :=*/ node.(type) {
	case *ast.Bundle, *ast.Body:
		fmt.Fprint(w, "}\n")

	case *ast.Promiser:
		fmt.Fprint(w, ";\n")

	case *ast.Function:
		fmt.Fprint(w, ")")

	case *ast.List:
		fmt.Fprint(w, " }")

	case *ast.ArgList:
		fmt.Fprint(w, ")\n{")

	case *ast.Constraint:
		if !last {
			fmt.Fprint(w, ",\n")
		}
	case *ast.Selection:
		fmt.Fprint(w, ";")
		if last {
			fmt.Fprint(w, "\n")
		}
	}
}

// print and set node to empty
func printChildrenOfType(w io.Writer, node ast.Node, format, typename string) {
	cs := node.Children()
	remove := []int{}
	for i, child := range cs {
		if reflect.TypeOf(child).String() == typename {
			fmt.Fprintf(w, format, child.Token().Lit)
			remove = append(remove, i)
		}
	}
	// reverse when removing, otherwise the indexes are broken.
	for i := len(remove) - 1; i >= 0; i-- {
		ast.Remove(node, remove[i])
	}
}
