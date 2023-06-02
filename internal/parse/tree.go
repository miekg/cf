package parse

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf/internal/rd"
)

type printFunc func(string)

func printChildrenOfType(w io.Writer, t *rd.Tree, tt chroma.TokenType, f printFunc) {
	detach := []*rd.Tree{}
	for _, c := range t.Subtrees {
		if ct, ok := c.Data().(chroma.Token); ok {
			if ct.Type == tt {
				f(ct.Value)
				detach = append(detach, c)
			}
		}
	}
	for i := range detach {
		t.Detach(detach[i])
	}
}

func printFirstChildOfType(w io.Writer, t *rd.Tree, tt chroma.TokenType, f printFunc) {
	for _, c := range t.Subtrees {
		if ct, ok := c.Data().(chroma.Token); ok {
			if ct.Type == tt {
				f(ct.Value)
				t.Detach(c)
				return
			}
		}
	}
}

func countOfType(parent *rd.Tree, t string) (i int) {
	for _, c := range parent.Subtrees {
		s, ok := c.Data().(string)
		if !ok {
			continue
		}
		if s == t {
			i++
		}
	}
	return i
}

// sequenceOfChild returns the index of the child in the parent, 0 is first, etc. Returns -1 when nothing is found.
func sequenceOfChild(parent, child *rd.Tree) int {
	for i, c := range parent.Subtrees {
		if c == child {
			return i
		}
	}
	return -1
}

// is the previous sibling of parent of type t.
func prevOfType(parent, child *rd.Tree, t string) bool {
	var p *rd.Tree
	for _, c := range parent.Subtrees {
		if c == child {
			if p == nil {
				return false
			}
			if s, ok := p.Data().(string); ok {
				if s == t {
					return true
				}
			}
		}
		p = c
	}
	return false
}

func firstOfType(parent, child *rd.Tree, t string) bool {
	return ofType(parent.Subtrees, child, t)
}

func lastOfType(parent, child *rd.Tree, t string) bool {
	rev := reverse(parent.Subtrees)
	return ofType(rev, child, t)
}

func ofType(children []*rd.Tree, child *rd.Tree, t string) bool {
	for _, c := range children {
		s, ok := c.Data().(string)
		if !ok {
			continue
		}
		if s == t {
			return c == child
		}
	}
	return false
}

func reverse(tree []*rd.Tree) []*rd.Tree {
	rev := make([]*rd.Tree, len(tree))

	j := 0
	for i := len(tree) - 1; i >= 0; i-- {
		rev[j] = tree[i]
		j++
	}
	return rev
}
