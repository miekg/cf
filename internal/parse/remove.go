package parse

import "github.com/shivamMg/rd"

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
	// only look at the top level and iterarte of the children ourselves.
	// the ast isn't perfect here, but it is a promiseguard followed by an empty
	// ClassPromise
	//      ├─ PromiseGuard
	//      │  └─ {KeywordDeclaration commands}
	//      ├─ ClassPromises
	//      ├─ PromiseGuard
	//      │  └─ {KeywordDeclaration files}
	//      └─ ClassPromises
	bb, ok := tree.Data().(string)
	if !ok {
		return
	}
	if bb != "BundleBody" {
		return
	}
	var pg *rd.Tree
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
			println(len(c.Subtrees))
			if len(c.Subtrees) == 0 {
				tree.Detach(pg)
			}
		}
	}
}
