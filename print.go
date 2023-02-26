package cf

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/miekg/cf/ast"
)

// Print pretty prints the CFengine AST in doc.
func Print(dst io.Writer, doc ast.Node) {
	for i, c := range doc.Children() {
		printRecur(dst, c, -1, i == 0, i == len(doc.Children())-1) // -1 because Specification is the top-level (noop) container.
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

	for _, c := range node.Token().Comment {
		fmt.Fprintf(w, "%s%s\n", indent, c)
	}

	// On Enter
	switch v := node.(type) {
	case *ast.Specification: // start of the tree, ignore

	case *ast.Bundle, *ast.Body:
		if !first && len(v.Token().Comment) == 0 {
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
		fmt.Fprintf(w, "\n%s%s\n", indent, v.Token().Lit)

	case *ast.ClassGuard:
		if !first {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s%s\n", indent, v.Token().Lit)

	case *ast.Promiser:
		children := len(v.Children()) != 0
		newline := ""
		if children {
			newline = "\n"
		}
		fmt.Fprintf(w, "%s%s%s", indent, v.Token().Lit, newline)

	case *ast.Constraint:
		fmt.Fprintf(w, "%s%s", indent, v.Token().Lit)

	case *ast.Selection:
		fmt.Fprintf(w, "\n%s%s", indent, v.Token().Lit)

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

	case *ast.Qstring, *ast.Identifier:
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
