package parse

import "github.com/miekg/cf/internal/rd"

func remove(tree *rd.Tree) {
	tvf := TreeVisitorFunc(func(tree *rd.Tree, entering bool) WalkStatus {
		if entering {
			removeEmptyPromiseGuards(tree) // detach empty promise guard.
		}
		return GoToNext
	})
	Walk(tree, tvf)
}

func removeEmptyPromiseGuards(tree *rd.Tree) {
	// only look at the top level and iterate of the children ourselves.
	// the ast isn't perfect here, but it is a promiseguard followed by an empty
	// ClassPromise
	//      ├─ PromiseGuard
	//      │  └─ {KeywordDeclaration commands}
	//      ├─ ClassPromises
	//      ├─ PromiseGuard
	//      │  └─ {KeywordDeclaration files}
	//      └─ ClassPromises
	// Save those up in `detach` and detach those tree in reverse order.
	bb, ok := tree.Data().(string)
	if !ok {
		return
	}
	if bb != "BundleBody" {
		return
	}
	var pg *rd.Tree
	detach := []*rd.Tree{} // save trees to detach, so we're not detaching while in the range-loop
	for _, c := range tree.Subtrees {
		c1, ok := c.Data().(string)
		if !ok {
			continue
		}
		// PromiseGuard / ClassPromises
		if c1 != "PromiseGuard" && c1 != "ClassPromises" {
			continue
		}

		switch c1 {
		case "PromiseGuard":
			pg = c
		case "ClassPromises":
			if len(c.Subtrees) == 0 {
				detach = append([]*rd.Tree{c}, detach...)
				detach = append([]*rd.Tree{pg}, detach...)
			}
		}
	}
	for _, d := range detach {
		tree.Detach(d)
	}
}
