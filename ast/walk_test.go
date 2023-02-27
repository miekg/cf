package ast

import (
	"testing"
)

func TestWalk(t *testing.T) {
	doc := &Specification{}
	Append(doc, &Identifier{})
	Append(doc, &Qstring{})

	i := 0
	nvf := NodeVisitorFunc(func(node Node, entering bool) WalkStatus {
		i++
		return GoToNext
	})
	Walk(doc, nvf)
	if i != 4 {
		t.Errorf("expected 4 for entering/leaving 1 container + 2 leafs, got %d", i)
	}
}
