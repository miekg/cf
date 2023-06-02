package parse

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
)

func constraintPreventSingleLine(constraint *rd.Tree) bool {
	for _, c := range constraint.Subtrees {
		ct, ok := c.Data().(chroma.Token)
		if !ok {
			return false
		}
		if ct.Type == chroma.KeywordReserved {
			switch ct.Value {
			case "contain":
				return true
			case "comment":
				return true
			}
			return false
		}
	}
	return false
}

func align(tree *rd.Tree) {
	tvf := TreeVisitorFunc(func(tree *rd.Tree, entering bool) WalkStatus {
		if entering {
			alignConstraints(tree)       // align on '=>' for multiple constraints
			alignPromisers(tree)         // align promises that have single constraints
			alignPromisersKeywords(tree) // align multiple constraints keywords, slist, string etc.
			alignSelections(tree)        // align on '=>' for multiple selection (body)
		}
		return GoToNext
	})
	Walk(tree, tvf)
}

func indentMultilineQstring(s, indent string) string {
	// this assumes we can remove any prefixing whitespace from succesive lines. TODO(miek): unsure if true
	lines := strings.Split(s, "\n") // Unix only now. TODO(miek)
	if len(lines) == 1 {
		return s
	}
	// Add newlines except for the last line, trim space, but put single one
	// back in, so it align under the opening quote.
	s = lines[0] + "\n"
	for i := 1; i < len(lines)-1; i++ {
		s += indent + " " + strings.TrimLeft(lines[i], "\t ") + "\n"
	}
	s += indent + " " + strings.TrimLeft(lines[len(lines)-1], "\t ")
	return s
}

// indentMultilineComments indent comments that have multiple lines, single line strings are left alone.
// when indentation is added, it is done for all lines except the first one.
func indentMultilineComment(s, indent string) string {
	if strings.Count(s, "\n") == 1 {
		return s
	}
	lines := strings.Split(s, "\n") // Unix only now. TODO(miek)
	s = lines[0] + "\n"
	for i := 1; i < len(lines); i++ {
		s += indent + strings.TrimLeft(lines[i], "\t ") + "\n"
	}
	return s
}

func alignConstraints(tree *rd.Tree) {
	promise, ok := tree.Data().(string)
	if !ok {
		return
	}
	if promise != "Promise" {
		return
	}

	max := -1
	align := []*rd.Tree{}
	for _, c := range tree.Subtrees {
		if len(c.Subtrees) < 1 {
			continue
		}

		constraint, ok := c.Data().(string)
		if !ok {
			continue
		}
		if constraint != "Constraint" {
			continue
		}
		token, ok := c.Subtrees[0].Data().(chroma.Token)
		if !ok {
			continue
		}
		if l := len(token.Value); l > max {
			max = l
		}
		align = append(align, c)
	}
	pad(align, max)
}

func alignPromisers(tree *rd.Tree) {
	max := -1
	align := []*rd.Tree{}
	for _, c := range tree.Subtrees {
		if len(c.Subtrees) < 1 {
			continue
		}
		promise, ok := c.Data().(string)
		if !ok {
			continue
		}
		if promise != "Promise" {
			continue
		}
		// this selects only single constraint promimises
		// still checks too many nodes
		if len(c.Subtrees) < 3 {
			continue
		}
		// check for comments in between. TODO(miek)
		token, ok := c.Subtrees[0].Data().(chroma.Token)
		if !ok {
			continue
		}
		if l := len(token.Value); l > max {
			max = l
		}
		align = append(align, c)
	}
	pad(align, max)
}

// alignPromisersKeywords aligns the string, slists in snippets like below, so that fat arrow is correctly
// aligned:
//
//	"path"     string => "etc/nftables.conf.d";
//	"acls"     slist  => getindices("acl");
//	"allacls"  slist  => lsdir("/etc/nftables.conf.d", ".*", "false");
func alignPromisersKeywords(tree *rd.Tree) {
	max := -1
	align := []*rd.Tree{}
	if len(tree.Subtrees) < 3 {
		return
	}
	for _, c := range tree.Subtrees {
		if len(c.Subtrees) < 1 {
			continue
		}
		promise, ok := c.Data().(string)
		if !ok {
			continue
		}
		if promise != "Promise" {
			continue
		}
		if len(c.Subtrees) < 3 {
			continue
		}
		if len(c.Subtrees[1].Subtrees) < 1 {
			continue
		}
		token, ok := c.Subtrees[1].Subtrees[0].Data().(chroma.Token)
		if !ok {
			continue
		}
		if token.Type != chroma.KeywordReserved {
			continue
		}
		if l := len(token.Value); l > max {
			max = l
		}
		align = append(align, c.Subtrees[1])
	}
	pad(align, max)
}

func alignSelections(tree *rd.Tree) {
	bodyselections, ok := tree.Data().(string)
	if !ok {
		return
	}
	if bodyselections != "BodySelections" {
		return
	}

	max := -1
	align := []*rd.Tree{}
	for _, c := range tree.Subtrees {
		if len(c.Subtrees) < 1 {
			continue
		}
		selection, ok := c.Data().(string)
		if !ok {
			continue
		}
		if selection != "Selection" {
			continue
		}
		token, ok := c.Subtrees[0].Data().(chroma.Token)
		if !ok {
			continue
		}
		if l := len(token.Value); l > max {
			max = l
		}
		align = append(align, c)
	}
	pad(align, max)
}

func pad(trees []*rd.Tree, max int) {
	if max == -1 {
		return
	}
	for _, t := range trees {
		token := t.Subtrees[0].Symbol.(chroma.Token)
		pad := max - len(token.Value)
		token.Value += strings.Repeat(" ", pad)
		t.Subtrees[0].Symbol = chroma.Token{Type: token.Type, Value: token.Value}
	}
}

// promisersAllHaveSingleConstraint checks if all Promise types have only a single constraint, this allows us to print them
// without newlines in between. This takes into account the contain and comment exception.
func promisersAllHaveSingleConstraint(tree *rd.Tree) bool {
	for _, c := range tree.Subtrees {
		promise, ok := c.Data().(string)
		if !ok {
			continue
		}
		if promise != "Promise" {
			continue
		}

		// now check the constraints
		constraints := 0
		for _, c1 := range c.Subtrees {
			constraint, ok := c1.Data().(string)
			if !ok {
				continue
			}
			if constraint != "Constraint" {
				continue
			}
			if constraintPreventSingleLine(c1) {
				constraints++ // fake +1 to signal 'do not compress'
			}
			constraints++
		}
		if constraints > 1 {
			return false
		}
	}
	return true
}
