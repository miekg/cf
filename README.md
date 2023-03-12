# CFEngine pretty printer

Cf is a formatter for CFEngine files, think of it as 'gofmt' for .cf files.

Cf can handle most CFEngine files.

- macros (@if etc, should also not be too hard)
- new `promise` keyword in CFEngine3 (should also not be too hard)

If a file has a top-level comment of the form: `# cffmt:no` the file will not be parsed and the
original input will be outputted instead.

If you have a "normal" looking CFEngine file that isn't parsed correctly, please open an issue with
the _most_ _minimal_ CFEngine syntax that fails to parse.

Cf aligns fat-arrows in a constraint, this is also true for selections in bodies.

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
aligned:

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
line. This is to show important things on the left hand side. (See align.go for details).

Trailing commas of lists are removed.

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

## Autofmt in (neo)vim

~~~
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin") " fmt
au BufWritePost *.cf silent call Fmt("cffmt /dev/stdin") " fmt on save
~~~

## Developing

Lexing is via Chroma (not 100% perfect, but I intent to upstream some changes there). We have a
recursive descent parser to create the AST, this us using *rd.Builder. Once we have the AST the
printing is relatively simple (`internal/parse/print.go`).

https://github.com/cfengine/core/blob/master/libpromises/cf3parse.y contains the grammar we're
reimplementing here.
