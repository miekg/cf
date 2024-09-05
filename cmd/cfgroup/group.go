package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/miekg/cf"
	"github.com/miekg/cf/internal/rd"
	"go.science.ru.nl/log"
)

// Group represents a CFengine group. The key in the map is the groupname and the slice are it's members.
type Groups map[string][]string

// Parse parses files and returns a Groups. On error the execution is log.Fatal-ed.
func Parse(files []string) Groups {
	var (
		err    error
		buffer []byte
		tree   *rd.Tree
		debug  *rd.DebugTree
	)
	groups := Groups{}
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		log.Debugf("Parsing %s", f)
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
			return groups
		}

		tree, debug, err = cf.ParseTokens(tokens)
		if err != nil {
			log.Fatalf("Can not parse %s: %s", f, err)
		}
		if tree == nil && debug == nil {
			log.Fatalf("Can not parse %s", f)
		}
		g1 := List(tree)
		groups = groups.Merge(g1)
	}
	return groups
}

// Names returns all group names (sorted).
func (g Groups) Names() []string {
	names := []string{}
	for k, _ := range g {
		names = append(names, k)
	}

	sort.Strings(names)
	return names
}

// Members returns the members for the groups in names. Non existing group names are ignored.
func (g Groups) Members(names []string) []string {
	members := []string{}
	for _, n := range names {
		members = append(members, g.Resolve(n)...)
	}
	sort.Strings(members)
	return members
}

// Resolve takes a group name and checks if any of the members exist in g, if so those are also resolved. The list
// returns contains the members of name that are leafs in g.
func (g Groups) Resolve(name string) []string {
	members := []string{}
	for _, m := range g[name] {
		_, ok := g[m]
		// if m does not exist in g, it is a leaf, add it.
		if !ok {
			members = append(members, m)
			continue
		}
		// m does exist as a group name: rescurse
		members = append(members, g.Resolve(m)...)
	}
	return members
}

// Search will search g for the (host)name, it returns all groups where this name is used, either directly or
// indirectly.
func (g Groups) Search(name string) []string {
	groups := []string{}

	for group, members := range g {
		for _, m := range members {
			if m == name {
				// direct member, add it
				groups = append(groups, group)
				continue
			}
			// might be indirect member
			indirect := g.Resolve(m)
			for _, im := range indirect {
				if im == name {
					groups = append(groups, group)
				}
			}
		}

	}

	sort.Strings(groups)
	return groups
}

// Merge merges two groups into one.
func (g Groups) Merge(g1 Groups) Groups {
	for k, members := range g1 {
		g[k] = append(g[k], members...)
	}
	return g
}

// Group is a single group, mostly used for building a Groups.
type Group struct {
	Name    string
	Members []string
}

// Print prints each element of sx on a single line to w. Duplicate elements are suppressed and the list is assumed
// to be sorted so duplicates can be suppressed.
func Print(w io.Writer, sx []string) {
	prev := ""
	for _, s := range sx {
		if s == prev {
			continue
		}
		fmt.Fprintln(w, s)
		prev = s
	}
}
