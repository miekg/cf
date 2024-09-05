package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/miekg/cf"
	"github.com/miekg/cf/internal/parse"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
	"go.science.ru.nl/log"
)

var (
	flagList    = flag.Bool("l", false, "list all defined groups")
	flagFiles   = flag.String("i", "", "comma seperated list of files to parse")
	flagReverse = flag.String("r", "", "show the classes for this specific host")
	flagDebug   = flag.Bool("d", false, "enable debug logging")
)

const (
	Groupcf       = "groups.cf"
	Promisescf    = "promises.cf"
	Functionalscf = "functionals.cf"
	Schedulecf    = "schedule.cf" // don't want this - remove
)

var Filescf = []string{Groupcf, Promisescf, Functionalscf, Schedulecf}

func main() {
	flag.Parse()
	var (
		err    error
		buffer []byte
	)

	if *flagDebug {
		log.D.Set()
	}

	cffiles := IsGit()
	log.Debugf("Using %v unless -i is given", cffiles)

	// TODO(miek): do something with stdin?

	var (
		tree   *rd.Tree
		debug  *rd.DebugTree
		groups Groups
	)

	files := strings.Split(*flagFiles, ",")

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
		if err != nil {
			log.Fatalf("Can not parse %s: %s", f, err)
		}
		if tree == nil && debug == nil {
			log.Fatalf("Can not parse %s", f)
		}
		groups = List(tree)
	}

	if *flagList {
		Print(os.Stdout, groups.Names())
		return
	}

	if *flagReverse != "" {
		Print(os.Stdout, groups.Search(*flagReverse))
		return
	}

	// No options, expect at least a group.
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

// IsGit returns the list of files of interest if the current cwd sits in a cfengine repository.
func IsGit() []string {
	paths := make([]string, len(Filescf))

	ctx := context.TODO()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Env = []string{"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null"}

	buf, err := cmd.CombinedOutput()
	if err != nil {
		// not a git repo
		for i, f := range Filescf {
			paths[i] = path.Join("/var/cfengine", f)
		}
		return paths
	}
	gitrepo := strings.TrimSpace(string(buf))
	// out should be a single line, that is the path of the git repo, check if the basename is 'cfengine'
	base := path.Base(gitrepo)
	prefix := "/var/cfengine"
	if base == "cfengine" {
		prefix = gitrepo
	}
	for i, f := range Filescf {
		paths[i] = path.Join(prefix, f)
	}
	return paths
}
