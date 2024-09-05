package main

import (
	"fmt"
	"io"
	"sort"
)

// Group represents a CFengine group. The key in the map is the groupname and the slice are it's members.
type Groups map[string][]string

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

	return groups
}

// Group is a single group, mostly used for building a Groups.
type Group struct {
	Name    string
	Members []string
}

// Print prints each element of sx on a single line to w. Duplicate elements are suppressed and the list is assumed
// to be sorted.
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
