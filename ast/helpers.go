package ast

import (
	"reflect"
)

// Up finds the root parent of n.
func Up(n Node) Node {
	for {
		if n.Parent() == nil {
			return n
		}
		n = n.Parent()
	}
}

// UpTo finds the type of to Node in the tree from n up to root. Each of the sibling in the current node are also
// checked.
func UpTo(n, to Node) Node {
	if reflect.TypeOf(n) == reflect.TypeOf(to) {
		return n
	}
	for {
		if n.Parent() == nil {
			return nil
		}
		for _, sibling := range n.Parent().Children() { // all our sibling
			if reflect.TypeOf(sibling) == reflect.TypeOf(to) {
				return n
			}
		}
		// exhausted current node, move one up
		n = n.Parent()
		if reflect.TypeOf(n) == reflect.TypeOf(to) {
			return n
		}
	}
}
