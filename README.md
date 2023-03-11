# CFEngine pretty printer

Cf is a formatter for CFEngine files, think of it as 'gofmt' for .cf files.

Cf can handle most CFEngine files, a few files I found that aren't parseable are stored in the
'unparseable' directory.

Current exceptions in parsing:

- comments in a list that is spread across multiple lines isn't handled
- qstring with quotes and backticks _and_ having # characters in them
- thin arrows (easily supported)

Cf aligns fat-arrows in a constraint:


~~~ cf
"/etc/apparmor.d"
             delete => tidy,
 	depth_search => recurse("0"),
             file_select => by_name("session");
~~~

becomes:

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

becomes:

~~~ cfengine
"getcapExists"  expression => fileexists("/sbin/getcap");
~~~

If there are multiple promises and they all have single constraints, the promises themselves are
aligned:

~~~ cfengine
"getcapExists"
     expression => fileexists("/sbin/getcap");

"setcapExists"  expression => fileexists("/sbin/setcap");
~~~

to:

~~~ cfengine
"getcapExists" expression => fileexists("/sbin/getcap");

"setcapExists" expression => fileexists("/sbin/setcap");
~~~

If comments are interspersed among these promises, the previous alignment stops. This makes comments
a natural barrier.

If a single constraint has a 'contain =>' or 'comment =>' they will _not_ be printed on the same
line. This is to show important things on the left hand side. (See align.go for details).

Trailing commas of lists are removed.

Package cf uses the lexer and parser from CFEngine's source and converts it into a (Go) AST that we
can walk and print. The AST is also exported and available to consumers of this package.

Install the `cffmt` binary with: `go install github.com/miekg/cf/cmd/cffmt@main`. Then use it by
giving it a filename or piping to standard input. The pretty printed document is printed to standard
output.

    ./cffmt ../../testdata/promtest.cf

Notes that cf will _not correctly parse_:

- Comments that are placed in a bundle/body but at the end. These will be dropped.
- Multiline comments with an _escaped_ quoting characters.
- Likely doesn't work with Windows line endings.
- Macros _are not parsed at all_.

## TODO

- Add tests with malformed content.
- promise guards don't have classguards as children, and they should.

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

This shows the following. The left side number is the number of spaces for the indentation (to
easily identify if nodes are on the same level).

## Autofmt in (neo)vim

~~~
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin") " fmt
au BufWritePost *.cf silent call Fmt("cffmt /dev/stdin") " fmt on save
~~~

## Developing

Lexing is via Chroma (not 100% perfect, but I intent to upstream some changes there). We have a
recursive decsent parser to create the AST, this us using *rd.Builder. Once we have the AST the
printing is relatively simple (`internal/parse/print.go`).
