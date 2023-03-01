package cf

import (
	"strings"

	"github.com/miekg/cf/ast"
)

// These contraint's are prevented from being put on a single line, even if there are the only child.
var preventSingleLine = []string{"contain", "comment"}

// constraintPreventSingleLine looks at the children of promiser and if only 1 _and_ contains a preventSingleLine
// keyword return true.
func constraintPreventSingleLine(promiser ast.Node) bool {
	cs := promiser.Children()
	if len(cs) != 1 {
		return false
	}
	constraint, ok := cs[0].(*ast.Constraint)
	if !ok {
		return false
	}
	for _, w := range preventSingleLine {
		if constraint.Token().Lit == w {
			return true
		}
	}
	return false
}

// align walks the Spec and for all Promisers with more than one Constraint, will align the contraint text
// in such a way the fat arrow (=>) align.
func align(doc ast.Node) {
	nvf := ast.NodeVisitorFunc(func(node ast.Node, entering bool) ast.WalkStatus {
		alignContraints(node) // align on '=>' for multiple constraints
		alignSelections(node) // align on '=>' for multiple selection (body)
		alignPromisers(node)  // align promises that have single constraints
		return ast.GoToNext
	})
	ast.Walk(doc, nvf)
}

func alignContraints(node ast.Node) {
	_, ok := node.(*ast.Promiser)
	if !ok {
		return
	}

	if len(node.Children()) == 1 {
		// only a single child, we will print this one 1 line, so there is no need to
		// align fat arrow, we do need to align all contraints them selves so it looks nice
		return
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
		return
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
}

func alignSelections(node ast.Node) {
	_, ok1 := node.(*ast.Body)
	_, ok2 := node.(*ast.ClassGuard)
	if !ok1 && !ok2 {
		return
	}
	max := -1
	for _, c := range node.Children() {
		_, ok := c.(*ast.Selection)
		if !ok {
			continue
		}
		if l := len(c.Token().Lit); l > max {
			max = l
		}
	}
	if max == -1 {
		return
	}
	// if still here, range over the node again and pad each contraint entry op to max.
	for _, c := range node.Children() {
		_, ok := c.(*ast.Selection)
		if !ok {
			continue
		}
		pad := max - len(c.Token().Lit)
		// c.Token() doesn't return a pointer, so use this roundabout way.
		t := c.Token()
		t.Lit += strings.Repeat(" ", pad)
		c.ResetToken(t)
	}
}

func alignPromisers(node ast.Node) {
	// don't care which nodes has promises, just align them if they are direct children.
	max := -1
	for _, c := range node.Children() {
		_, ok := c.(*ast.Promiser)
		if !ok {
			continue
		}
		if l := len(c.Token().Lit); l > max {
			max = l
		}
	}
	if max == -1 {
		return
	}
	for _, c := range node.Children() {
		_, ok := c.(*ast.Promiser)
		if !ok {
			continue
		}
		pad := max - len(c.Token().Lit)
		t := c.Token()
		t.Lit += strings.Repeat(" ", pad)
		c.ResetToken(t)
	}
}
