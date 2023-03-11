package parse

import "github.com/shivamMg/rd"

type WalkStatus int

const (
	// GoToNext is the default traversal of every node.
	GoToNext WalkStatus = iota
	// SkipChildren tells walker to skip all children of current node.
	SkipChildren
	// Terminate tells walker to terminate the traversal.
	Terminate
)

type TreeVisitor interface {
	Visit(tree *rd.Tree, entering bool) WalkStatus
}

// TreeVisitorFunc casts a function to match NodeVisitor interface.
type TreeVisitorFunc func(tree *rd.Tree, entering bool) WalkStatus

// Walk traverses tree recursively.
func Walk(tree *rd.Tree, visitor TreeVisitor) WalkStatus {

	status := visitor.Visit(tree, true) // entering
	if status == Terminate {
		visitor.Visit(tree, false)
		return status
	}
	if status != SkipChildren {
		for _, t := range tree.Subtrees {
			status = Walk(t, visitor)
			if status == Terminate {
				return status
			}
		}
	}
	status = visitor.Visit(tree, false) // exiting
	if status == Terminate {
		return status
	}
	return GoToNext
}

// Visit calls visitor function.
func (f TreeVisitorFunc) Visit(tree *rd.Tree, entering bool) WalkStatus { return f(tree, entering) }

// WalkFunc is like Walk but accepts just a callback function.
func WalkFunc(tree *rd.Tree, f TreeVisitorFunc) {
	visitor := TreeVisitorFunc(f)
	Walk(tree, visitor)
}
