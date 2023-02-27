package ast

import (
	"testing"
)

func TestRemove(t *testing.T) {
	doc := &Specification{}
	Append(doc, &Identifier{})
	Append(doc, &Qstring{})

	Reverse(doc)
	cs := doc.Children()
	if v, ok := cs[0].(*Qstring); !ok {
		t.Errorf("expecting qstring, got %T", v)
	}
	if v, ok := cs[1].(*Identifier); !ok {
		t.Errorf("expecting qstring, got %T", v)
	}
}
