package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/miekg/cf/internal/parse"
	"github.com/miekg/cf/internal/rd"
	"github.com/miekg/cf/internal/token"
	"go.science.ru.nl/log"
)

var (
	flagList    = flag.Bool("l", false, "list all defined groups")
	flagDebug   = flag.Bool("d", false, "enable debug logging")
	flagOnce    = flag.Bool("o", false, "list hosts that are only used once")
	flagMore    = flag.Bool("n", false, "list hosts that are used more than once")
	flagFiles   = flag.String("i", "", "comma seperated list of files to parse")
	flagReverse = flag.String("r", "", "show the groups/classes for this specific host")
	flagNot     = flag.String("x", "", "list hosts that are in GROUP, but not in this group")
	flagVersion = flag.Bool("v", false, "show version")
)

const (
	Prefixcf      = "/masterfiles/adm/"
	Groupcf       = Prefixcf + "groups.cf"
	Promisescf    = Prefixcf + "promises.cf"
	Functionalscf = Prefixcf + "functionals.cf"
)

var version = "n/a"

var Filescf = []string{Groupcf, Promisescf, Functionalscf}

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *flagDebug {
		log.D.Set()
	}

	groups := Groups{}
	var err error
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		log.Debug("Using standard input")
		if groups, err = Parse([]io.Reader{os.Stdin}); err != nil {
			log.Fatalf("Failed to parse %s: %s", "standard input", err)
		}
	} else {
		files := IsGit()
		log.Debugf("Using %v", files)
		rs := make([]io.Reader, len(files))
		for i, f := range files {
			if rs[i], err = os.Open(f); err != nil {
				log.Fatalf("Failed to open %s: %s", f, err)
			}
		}
		if groups, err = Parse(rs); err != nil {
			log.Fatalf("Failed to parse file: %s", err)
		}
	}

	if *flagList {
		Print(os.Stdout, groups.Names())
		return
	}

	if *flagReverse != "" {
		Print(os.Stdout, groups.Search(*flagReverse))
		return
	}
	if *flagNot != "" {
		exclude := groups.Members([]string{*flagNot})
		all := groups.Members(flag.Args())
		seen := map[string]struct{}{}
		// remove exclude from all
		for _, a := range all {
			seen[a] = struct{}{}
		}
		for _, e := range exclude {
			delete(seen, e)
		}
		// re-assemble to slice, so we can use Print
		members := []string{}
		for k := range seen {
			members = append(members, k)
		}
		sort.Strings(members)
		Print(os.Stdout, members)
		return
	}

	if *flagOnce {
		members := groups.Members(flag.Args())
		seen := map[string]int{}
		for _, m := range members {
			seen[m]++
		}
		prev := ""
		for _, m := range members {
			if seen[m] != 1 {
				continue
			}
			if m == prev {
				continue
			}
			fmt.Fprintln(os.Stdout, m)
			prev = m
		}
		return
	}
	if *flagMore {
		members := groups.Members(flag.Args())
		seen := map[string]int{}
		for _, m := range members {
			seen[m]++
		}
		prev := ""
		for _, m := range members {
			if seen[m] < 2 {
				continue
			}
			if m == prev {
				continue
			}
			fmt.Fprintln(os.Stdout, m)
			prev = m
		}
		return
	}

	// No options, expect at least a group.
	Print(os.Stdout, groups.Members(flag.Args()))
}

/*
List walks the ast and returns the "members" of each promise. This only works is it's fed groups.cf and friends.

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
						if l != token.List {
							continue
						}
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
