package cf

import (
	"fmt"
	"testing"

	"github.com/miekg/cf/internal/parse"
)

func TestSnippetSelection(t *testing.T) {
	tests := []tc{
		{`bundlesequence => { "hello_world" };`, `Selection
├─ {KeywordReserved bundlesequence}
├─ FatArrow
│  └─ {TokenType(-996) =>}
├─ Rval
│  └─ List
│     └─ Litem
│        └─ Qstring
│           └─ {TokenType(-994) "hello_world"}
└─ {Punctuation ;}
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.Selection) })
	}
}

func TestSnippetSelections(t *testing.T) {
	tests := []tc{
		{`mode => "644";
		  groups => "root";`, `BodySelections
├─ Selection
│  ├─ {KeywordReserved mode}
│  ├─ FatArrow
│  │  └─ {TokenType(-996) =>}
│  ├─ Rval
│  │  └─ Qstring
│  │     └─ {TokenType(-994) "644"}
│  └─ {Punctuation ;}
└─ Selection
   ├─ {KeywordReserved groups}
   ├─ FatArrow
   │  └─ {TokenType(-996) =>}
   ├─ Rval
   │  └─ Qstring
   │     └─ {TokenType(-994) "root"}
   └─ {Punctuation ;}
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.BodySelections) })
	}
}

func TestSnippetClassGuardSelections(t *testing.T) {
	tests := []tc{
		{` Any::
		    mode => "644";
                    owners => { "root" };
                    groups => { "root" };`, `ClassGuardSelections
├─ {NameClass Any}
└─ BodySelections
   ├─ Selection
   │  ├─ {KeywordReserved mode}
   │  ├─ FatArrow
   │  │  └─ {TokenType(-996) =>}
   │  ├─ Rval
   │  │  └─ Qstring
   │  │     └─ {TokenType(-994) "644"}
   │  └─ {Punctuation ;}
   ├─ Selection
   │  ├─ {KeywordReserved owners}
   │  ├─ FatArrow
   │  │  └─ {TokenType(-996) =>}
   │  ├─ Rval
   │  │  └─ List
   │  │     └─ Litem
   │  │        └─ Qstring
   │  │           └─ {TokenType(-994) "root"}
   │  └─ {Punctuation ;}
   └─ Selection
      ├─ {KeywordReserved groups}
      ├─ FatArrow
      │  └─ {TokenType(-996) =>}
      ├─ Rval
      │  └─ List
      │     └─ Litem
      │        └─ Qstring
      │           └─ {TokenType(-994) "root"}
      └─ {Punctuation ;}
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) { doTest(t, i, test, parse.ClassGuardSelections) })
	}
}
