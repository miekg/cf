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

	return sort.StringSlice(names)
}

// Members returns the members for the groups in names. Non existing group names are ignored.
// FIXME(miek): this needs to resolve to final machines, if a name exist as a group, furhter resolve it.
func (g Groups) Members(names []string) []string {
	members := []string{}
	for _, n := range names {
		members = append(members, g[n]...)
	}
	return members
}

// Group is a single group, mostly used for building a Groups.
type Group struct {
	Name    string
	Members []string
}

// Print prints each element of sx on a single line to w.
func Print(w io.Writer, sx []string) {
	for _, s := range sx {
		fmt.Fprintln(w, s)
	}
}
