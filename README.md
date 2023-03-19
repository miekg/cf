# CFEngine pretty printer

Cf is a formatter for CFEngine files, think of it as 'gofmt' (from golang) for .cf files. See
cmd/cffmt for the CLI.

Cf should handle all CFEngine files, allthough the syntax is _so_ liberal, especially where you can
place comments (the official yacc parser/lexer throws away comments) that there is always a chance a
file isn't parseable. The new data type which can parse inline json can also cause trouble.

If a file has a top-level comment of the form: `# cffmt:no` the file will not be parsed and the
original input will be outputted instead.

If you have a "normal" looking CFEngine file that isn't parsed correctly, please open an issue with
the _most_ _minimal_ CFEngine syntax that fails to parse.

## Layout

Cf uses an indent of 2 spaces to indent deeper elements of the tree.

- the promise guard (i.e. `files:` has 2 newlines above it, if its not the first in the file
- the glass guard (i.e. `any::`), if given has a empty line above it, but is attached to the
  promiser.
- the promiser is always attached to the constraint expressions
- empty promise guards are removed, i.e. `commands:` without any commands defined will be removed
  from the output

Cf aligns fat-arrows in constraint expressions, this is also true for selections in bodies.

~~~ cf
"/etc/apparmor.d"
             delete => tidy,
 	depth_search => recurse("0"),
             file_select => by_name("session");
~~~

Becomes:

~~~ cfengine
"/etc/apparmor.d"
  delete       => tidy,
  depth_search => recurse("0"),
  file_select  => by_name("lightdm-guest-session");
~~~

If there is only a single constraint it will be printed on the same line:

~~~ cfengine
   "getcapExists"
        expression => fileexists("/sbin/getcap");
~~~

Becomes:

~~~ cfengine
"getcapExists" expression => fileexists("/sbin/getcap");
~~~

If there are multiple promises and they all have single constraints, the promises themselves are
aligned and the newline between them is scapped:

~~~ cfengine
"getcapExists"
     expression => fileexists("/sbin/getcap");

"setcapExists"  expression => fileexists("/sbin/setcap");
~~~

To:

~~~ cfengine
"getcapExists" expression => fileexists("/sbin/getcap");
"setcapExists" expression => fileexists("/sbin/setcap");
~~~

If a single constraint has a 'contain =>' or 'comment =>' they will _not_ be printed on the same
line. This is to show important things on the left hand side. (See align.go for details), i.e:

~~~ cfengine
printvm::
 "printer[xxx]"	string => "ps.ppd";
  "printer[xxx]"		string => "ps.ppd";
  "printer[xxx]" string => "ps.ppd";
~~~

To:

~~~ cfengine
printvm::
  "printer[xxx]" string => "ps.ppd";
  "printer[xxx]" string => "ps.ppd";
  "printer[xxx]" string => "ps.ppd";
~~~

But if one of the constraints was `contain` or `comment`:

~~~ cfengine
printvm::
  "printer[xxx]" string => "ps.ppd";
  "printer[xxx]" comment => "ps.ppd";
  "printer[xxx]" string => "ps.ppd";
~~~

Will instead become:
~~~ cfengine
printvm::
  "printer[xxx]" string => "ps.ppd";

  "printer[xxx]"
    comment => "ps.ppd";

  "printer[xxx]" string => "ps.ppd";
~~~

Trailing commas of lists are removed. List are wrapped at the 120th column: (assuming ggg, is on the
120th column):

~~~ cfengine
"Clients"         or => { aaa, bbb, ccc, dddd, eee, fff,
  ggg, hhhh };
~~~

To:

~~~ cfengine
"Clients"         or => { aaa, bbb, ccc, dddd, eee, fff, ggg,
                          hhhh };
~~~

It also makes sure there isn't a dangling `};` on a line.

Install the `cffmt` binary with: `go install github.com/miekg/cf/cmd/cffmt@main`. Then use it by
giving it a filename or piping to standard input. The pretty printed document is printed to standard
output.

    ./cffmt ../../testdata/promtest.cf

## Abstract Syntax Tree

If you only want see the AST use -a, and throw away standard output:

~~~
cmd/cffmt/cffmt -a -p=false testdata/arg-list.cf >/dev/null
2023/03/11 22:29:51 Parse Tree:
Specification
└─ Bundle
   ├─ {Keyword bundle}
   ├─ {Keyword agent}
   ├─ {NameFunction bla}
   └─ BundleBody
      ├─ PromiseGuard
      │  └─ {KeywordDeclaration vars}
      └─ ClassPromises
         └─ Promise
            ├─ {TokenType(-994) "installed_canonified"}
            ├─ Constraint
            │  ├─ {KeywordType slist}
            │  ├─ FatArrow
            │  │  └─ {TokenType(-996) =>}
            │  └─ Rval
            │     └─ Qstring
            │        └─ {TokenType(-994) "aaa"}
            └─ {Punctuation ;}
~~~

From this input file:

~~~ cfengine
bundle agent bla
{
 vars:
    "installed_canonified"
        slist => "aaa";
}
~~~

The plain strings, i.e. `Bundle` are non-terminals, while the `{TokenType(-994) ...}` and
`{KeywordTYpe ...}` are terminals. That first terminal has a "local" type that we define, in this
case it is a `token.Qstring`, otherwise it's a original `chroma.Token`. In both cases the type is a
`chroma.Token`. See `internal/parse/print.go` on how that tree is walked.

## Autofmt in (neo)vim

~~~
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin") " fmt
au BufWritePost *.cf silent call Fmt("cffmt /dev/stdin") " fmt on save
~~~

## Developing

Lexing is via Chroma (not 100% perfect, but we work around this in `lex.go`). We have a
recursive descent parser to create the AST, this is using *rd.Builder. Once we have the AST the
printing is relatively simple (`internal/parse/print.go`).

https://github.com/cfengine/core/blob/master/libpromises/cf3parse.y contains the grammar we're
reimplementing here. Note that one doesn't deal with comments, and is not used to build an AST.
