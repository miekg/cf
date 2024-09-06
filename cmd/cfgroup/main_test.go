package main

import (
	"io"
	"os"
	"sort"
	"testing"
)

// FIXME(miek): tests duplicated code, instead of calling a function.

func testgroups(t *testing.T) Groups {
	f, _ := os.Open("testdata/groups.cf")
	groups, _ := Parse([]io.Reader{f})
	return groups
}

func TestOnce(t *testing.T) {
	groups := testgroups(t)

	members := groups.Members([]string{"IsWebserver"})
	seen := map[string]int{}
	for _, m := range members {
		seen[m]++
	}
	prev := ""
	once := []string{}
	for _, m := range members {
		if seen[m] != 1 {
			continue
		}
		if m == prev {
			continue
		}
		once = append(once, m)
		prev = m
	}

	t.Logf("once %v", once)
	if len(once) != 1 {
		t.Fatalf("expected 1 element, got %d", len(once))
	}
	if once[0] != "webserver1" {
		t.Fatalf("expected %s to be the host, got %s", "webserver1", once[0])
	}
}

func TestExclude(t *testing.T) {
	groups := testgroups(t)

	exclude := groups.Members([]string{"IsInactive"})
	all := groups.Members([]string{"IsWebserver"})
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
	if len(members) != 1 {
		t.Fatalf("expected 1 element, got %d", len(members))
	}
	// webserver1 exists in inactive
	if members[0] != "webserver2" {
		t.Fatalf("expected %s to be the host, got %s", "webserver2", members[0])
	}
}
