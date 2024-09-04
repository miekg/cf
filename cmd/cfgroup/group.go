package main

import "sort"

// Group represents a CFengine group. The key in the map is the groupname and the slice are it's members.
type Groups map[string][]string

// Keys returns all group names (sorted).
func (g Groups) Keys() []string {
	names := []string{}
	for k, _ := range g {
		names = append(names, k)
	}

	return sort.StringSlice(names)
}

// Group is a single group, mostly used for building a Groups.
type Group struct {
	Name    string
	Members []string
}
