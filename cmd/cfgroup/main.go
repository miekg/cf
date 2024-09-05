package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/cf"
	"github.com/miekg/cf/internal/parse"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
)

var (
	flagList    = flag.Bool("l", false, "list all defined groups")
	flagFiles   = flag.String("i", "", "comma seperated list of files to parse")
	flagReverse = flag.String("r", "", "show the classes for this specific host")
)

func main() {
	flag.Parse()
	var (
		err    error
		buffer []byte
	)

	// implements groups on the commandline

	// FIXME(miek): do something with stdin?

	var (
		tree   *rd.Tree
		debug  *rd.DebugTree
		groups Groups
	)

	files := strings.Split(*flagFiles, ",")
	// FIXME(miek): get cfengine files from usual location ,
	// git rev-parse --show-toplevel
	// otherwise default location

	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		buffer, err = os.ReadFile(f)
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

		tree, debug, err = cf.ParseTokens(tokens)
		if tree == nil && debug == nil && err == nil {
			return
		}
		groups = List(tree)
	}

	if *flagList {
		Print(os.Stdout, groups.Names())
		return
	}

	// no options, expect a last on group
	Print(os.Stdout, groups.Members(flag.Args()))
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
func List(tree *rd.Tree) Groups {
	groups := Groups{}
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
			groups[g.Name] = g.Members
		}

		return parse.GoToNext
	})
	parse.Walk(tree, tvf)
	return groups
}
