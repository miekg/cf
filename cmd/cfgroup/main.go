package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/miekg/cf"
	"github.com/miekg/cf/internal/parse"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

var (
	flagList = flag.Bool("l", false, "list all defined groups")
)

func main() {
	flag.Parse()
	var (
		err    error
		buffer []byte
	)

	switch flag.NArg() {
	case 0:
		buffer, err = io.ReadAll(os.Stdin)
	case 1:
		buffer, err = os.ReadFile(flag.Arg(0))
	default:
		log.Fatal("Too many arguments")
	}
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := cf.Lex(string(buffer))
	if err != nil {
		log.Fatal(err)
	}
	if cf.IsNoParse(tokens) {
		fmt.Printf("%s", buffer)
		return
	}

	parseTree, debugTree, err := cf.ParseTokens(tokens)
	if parseTree == nil && debugTree == nil && err == nil {
		return
	}
	if *flagList {
		groups := List(parseTree)
		for i := range groups {
			fmt.Println(groups[i].Name)
		}
	}
}

/*
List walks the ast and returns the "members" of each promise. This only works is it's fed cfgroup.cf.

The ast looks like:

	├─ Promise
	│  ├─ {TokenType(-994) "ClusterMLP"}
	│  ├─ Constraint
	│  │  ├─ {KeywordReserved or}
	│  │  ├─ FatArrow
	│  │  │  └─ {TokenType(-996) =>}
	│  │  └─ Rval
	│  │     └─ List
	│  │        ├─ Litem
	│  │        │  └─ {NameFunction mlp01}
	│  │        ├─ {Punctuation ,}
	│  │        ├─ Litem
	│  │        │  └─ {NameFunction mlp02}
	│  │        ├─ {Punctuation ,}

For each Promise we want to items from the Rval after the FatArrow
*/
func List(tree *rd.Tree) []Group {
	groups := []Group{}
	// also check top level Class, so we only get "vars".
	tvf := parse.TreeVisitorFunc(func(tree *rd.Tree, entering bool) parse.WalkStatus {
		if !entering {
			return parse.GoToNext
		}
		t, ok := tree.Data().(string)
		if t != token.Promise || !ok {
			return parse.GoToNext
		}
		g := Group{}
		for i, s1 := range tree.Subtrees {
			if i == 0 {
				if t, ok := s1.Data().(token.T); ok {
					g.Name, _ = strconv.Unquote(t.Value)
				}
			}
			c, ok := s1.Data().(string)
			if !ok {
				continue
			}
			if c != token.Constraint {
				continue
			}

			for _, s2 := range s1.Subtrees {
				r, ok := s2.Data().(string)
				if !ok {
					continue
				}
				if r == token.Rval {
					for _, s3 := range s2.Subtrees {
						l, ok := s3.Data().(string)
						if !ok {
							continue
						}
						if l == token.List {
							for _, s4 := range s3.Subtrees {
								li, ok := s4.Data().(string)
								if !ok {
									continue
								}
								if li == token.Litem {
									// there should be one 1 item here, just grab it
									s5 := s4.Subtrees[0]
									if t, ok := s5.Data().(token.T); ok {
										g.Members = append(g.Members, t.Value)
									}
								}
							}
						}
					}
				}
			}
			groups = append(groups, g)
		}

		return parse.GoToNext
	})
	parse.Walk(tree, tvf)
	return groups
}
