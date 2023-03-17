package cf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/miekg/cf/internal/parse"
	"github.com/shivamMg/rd"
)

// Test short snipppets of cf syntax, without pulling in the whole tree
// These need to match the ast

type tc struct {
	input string
	ast   string
}

// astCompare compares the AST string in a with that in b.
func astCompare(a, b string) error {
	as := strings.Split(a, "\n")
	bs := strings.Split(b, "\n")
	for i := range as {
		if as[i] != bs[i] {
			return fmt.Errorf("line %d, doesn't match %s != %s", i, as[i], bs[i])
		}
	}
	return nil
}

func doTest(t *testing.T, i int, test tc, testfunc func(b *rd.Builder) bool) {
	b, tokens, err := newBuilder(test.input + "\n")
	if err != nil {
		t.Fatal(err)
	}

	testfunc(b)
	x := b.ParseTree()
	if x == nil {
		for i := range tokens {
			t.Logf("%v\n", tokens[i])
		}
		t.Log("Debug tree:\n\n", b.DebugTree())
		t.Errorf("Test %d failed to parse", i)
		return
	}
	if err := astCompare(test.ast, b.ParseTree().String()); err != nil {
		t.Errorf("Test %d, AST doesn't match: %s", i, err)
		t.Logf("Test %d, AST\n%s\n", i, b.ParseTree())
		t.Logf("Expect AST\n%s\n", test.ast)
		t.Log("Debug tree:\n\n", b.DebugTree())
	}
}

func TestSnippetConstraints(t *testing.T) {
	// Note these are bare contraints, without closing , or ;
	tests := []tc{
		{`inform => $(compounds.to_inform)`, `Constraint
├─ {KeywordReserved inform}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ NakedVar
      └─ {NameVariable $(compounds.to_inform)}
`,
		},
		{`slist => "hallo"`, `Constraint
├─ {KeywordReserved slist}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Qstring
      └─ {TokenType(-994) "hallo"}
`,
		},
		{`slist => maplist()`, `Constraint
├─ {KeywordReserved slist}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Function
      ├─ {NameFunction maplist}
      └─ GiveArgList
`,
		},
		{`container => in_shell`, `Constraint
├─ {KeywordReserved container}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ {NameFunction in_shell}
`,
		},
		{`perms => mog(0444, root, root)`, `Constraint
├─ {KeywordReserved perms}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Function
      ├─ {NameFunction mog}
      └─ GiveArgList
         ├─ GaItem
         │  └─ {LiteralNumberInteger 0444}
         ├─ {Punctuation ,}
         ├─ GaItem
         │  └─ {NameFunction root}
         ├─ {Punctuation ,}
         └─ GaItem
            └─ {NameFunction root}
`,
		},
		{`slist => maplist("aa", nog("bb"), "Monday")`, `Constraint
├─ {KeywordReserved slist}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Function
      ├─ {NameFunction maplist}
      └─ GiveArgList
         ├─ GaItem
         │  └─ Qstring
         │     └─ {TokenType(-994) "aa"}
         ├─ {Punctuation ,}
         ├─ GaItem
         │  └─ Function
         │     ├─ {NameFunction nog}
         │     └─ GiveArgList
         │        └─ GaItem
         │           └─ Qstring
         │              └─ {TokenType(-994) "bb"}
         ├─ {Punctuation ,}
         └─ GaItem
            └─ Qstring
               └─ {TokenType(-994) "Monday"}
`,
		},
		{`slist => maplist(canonify("$(this)"), @(installed_names))`, `Constraint
├─ {KeywordReserved slist}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Function
      ├─ {NameFunction maplist}
      └─ GiveArgList
         ├─ GaItem
         │  └─ Function
         │     ├─ {NameFunction canonify}
         │     └─ GiveArgList
         │        └─ GaItem
         │           └─ Qstring
         │              └─ {TokenType(-994) "$(this)"}
         ├─ {Punctuation ,}
         └─ GaItem
            └─ NakedVar
               └─ {NameVariable @(installed_names)}
`,
		},
		{`copy_from => no_backup_rdcp("$(def.distr_files_dir)/etc/update-motd.d/51-git-status", $(sys.policy_hub))`, `Constraint
├─ {KeywordReserved copy_from}
├─ FatArrow
│  └─ {TokenType(-996) =>}
└─ Rval
   └─ Function
      ├─ {NameFunction no_backup_rdcp}
      └─ GiveArgList
         ├─ GaItem
         │  └─ Qstring
         │     └─ {TokenType(-994) "$(def.distr_files_dir)/etc/update-motd.d/51-git-status"}
         ├─ {Punctuation ,}
         └─ GaItem
            └─ NakedVar
               └─ {NameVariable $(sys.policy_hub)}
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.Constraint) })
	}
}

func TestSnippetPromisees(t *testing.T) {
	tests := []tc{
		{
			`"Skipped self upgrade desired version $(sys.cf_version)" -> { "ENT-3592" };`, `Promise
├─ {TokenType(-994) "Skipped self upgrade desired version $(sys.cf_version)"}
├─ Promisee
│  ├─ ThinArrow
│  │  └─ {TokenType(-995) ->}
│  └─ Rval
│     └─ List
│        └─ Litem
│           └─ Qstring
│              └─ {TokenType(-994) "ENT-3592"}
└─ {Punctuation ;}
`,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.Promises) })
	}
}

func TestSnippetPromises(t *testing.T) {
	tests := []tc{
		{
			`"/etc/prometheus/prometheus.yml"
			perms => mog(0444, root, root),
			classes => if_repaired("prometheus_updated");`, `Promise
├─ {TokenType(-994) "/etc/prometheus/prometheus.yml"}
├─ Constraint
│  ├─ {KeywordReserved perms}
│  ├─ FatArrow
│  │  └─ {TokenType(-996) =>}
│  └─ Rval
│     └─ Function
│        ├─ {NameFunction mog}
│        └─ GiveArgList
│           ├─ GaItem
│           │  └─ {LiteralNumberInteger 0444}
│           ├─ {Punctuation ,}
│           ├─ GaItem
│           │  └─ {NameFunction root}
│           ├─ {Punctuation ,}
│           └─ GaItem
│              └─ {NameFunction root}
├─ {Punctuation ,}
├─ Constraint
│  ├─ {KeywordReserved classes}
│  ├─ FatArrow
│  │  └─ {TokenType(-996) =>}
│  └─ Rval
│     └─ Function
│        ├─ {NameFunction if_repaired}
│        └─ GiveArgList
│           └─ GaItem
│              └─ Qstring
│                 └─ {TokenType(-994) "prometheus_updated"}
└─ {Punctuation ;}
`,
		},
		{
			`"/etc/prometheus/prometheus.yml";`, `Promise
├─ {TokenType(-994) "/etc/prometheus/prometheus.yml"}
└─ {Punctuation ;}
`,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.Promises) })
	}

}

func TestSnippetClassPromises(t *testing.T) {
	tests := []tc{
		{
			`IsCupsServer.cupsd_updated::
			"/usr/sbin/service cups restart";`, `ClassPromises
└─ ClassGuardPromises
   ├─ {NameClass IsCupsServer.cupsd_updated}
   └─ Promise
      ├─ {TokenType(-994) "/usr/sbin/service cups restart"}
      └─ {Punctuation ;}
`,
		},
		{
			`IsCupsServer.cupsd_updated::
			"/usr/sbin/service cups restart";
			IsCupsServer.cupsd_updated::
			"/usr/sbin/service cups restart";`, `ClassPromises
├─ ClassGuardPromises
│  ├─ {NameClass IsCupsServer.cupsd_updated}
│  └─ Promise
│     ├─ {TokenType(-994) "/usr/sbin/service cups restart"}
│     └─ {Punctuation ;}
└─ ClassGuardPromises
   ├─ {NameClass IsCupsServer.cupsd_updated}
   └─ Promise
      ├─ {TokenType(-994) "/usr/sbin/service cups restart"}
      └─ {Punctuation ;}
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.ClassPromises) })
	}
}

func TestSnippetBundleBody(t *testing.T) {
	tests := []tc{
		{
			`files:

reports:

    "/usr/sbin/service cups restart";`, `BundleBody
├─ PromiseGuard
│  └─ {KeywordDeclaration files}
├─ ClassPromises
├─ PromiseGuard
│  └─ {KeywordDeclaration reports}
└─ ClassPromises
   └─ Promise
      ├─ {TokenType(-994) "/usr/sbin/service cups restart"}
      └─ {Punctuation ;}
`,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.BundleBody) })
	}
}

func newBuilder(buf string) (*rd.Builder, []rd.Token, error) {
	tokens, err := Lex(buf)
	if err != nil {
		return nil, tokens, err
	}
	return rd.NewBuilder(tokens), tokens, nil
}
