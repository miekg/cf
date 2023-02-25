package ast

// don't know if worth keeping, should use WalkFunc

import (
	"fmt"
	"io"
	"strings"
)

// Print pretty prints the CFengine AST in doc.
func Print(dst io.Writer, doc Node) {
	for _, c := range doc.Children() {
		printRecur(dst, c, "  ", 0)
	}
}

// older debug function, might be removed at some point.
func printDefault(w io.Writer, indent string, typeName string, token Token) {
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

func printRecur(w io.Writer, node Node, prefix string, depth int) {
	if node == nil {
		return
	}
	indent := strings.Repeat(prefix, depth)

	switch v := node.(type) {
	default:
		printDefault(w, indent, fmt.Sprintf("%T", v), v.Token())
	}
	for _, child := range node.Children() {
		printRecur(w, child, prefix, depth+1)
	}
}
