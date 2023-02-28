package cf

import (
	"strings"

	"github.com/miekg/cf/ast"
)

// fatArrowAlign walks the Spec and for all Promisers with more than one Constraint, will align the contraint text
// in such a way the fat arrow (=>) align.
func fatArrowAlign(doc ast.Node) {
	nvf := ast.NodeVisitorFunc(func(node ast.Node, entering bool) ast.WalkStatus {
		_, ok := node.(*ast.Promiser)
		if !ok {
			return ast.GoToNext
		}

		max := -1
		for _, c := range node.Children() {
			_, ok := c.(*ast.Constraint)
			if !ok {
				continue
			}
			if l := len(c.Token().Lit); l > max {
				max = l
			}
		}
		if max == -1 {
			return ast.GoToNext
		}
		// if still here, range over the node again and pad each contraint entry op to max.
		for _, c := range node.Children() {
			_, ok := c.(*ast.Constraint)
			if !ok {
				continue
			}
			pad := max - len(c.Token().Lit)
			// c.Token() doesn't return a pointer, so use this roundabout way.
			t := c.Token()
			t.Lit += strings.Repeat(" ", pad)
			c.ResetToken(t)
		}

		return ast.GoToNext
	})
	ast.Walk(doc, nvf)
}
