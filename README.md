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
cmd/cffmt/cffmt -a testdata/arglist.cf >/dev/null
~~~

This shows the following. The left side number is the number of spaces for the indentation (to
easily identify if nodes are on the same level).

~~~ txt
 0 *ast.Specification
 2   *ast.Bundle 'bundle'
 4     *ast.Identifier 'bla'
 4     *ast.Identifier 'bla'
 4     *ast.PromiseGuard 'vars:'
 6       *ast.Promiser '"installed_names_canonified"'
 8         *ast.Constraint 'slist'
10           *ast.FatArrow '=>'
10           *ast.Function 'maplist'
12             *ast.GiveArgItem
14               *ast.Function 'canonify'
16                 *ast.GiveArgItem
18                   *ast.Qstring '"$(this)"'
14               *ast.GiveArgItem
16                 *ast.NakedVar '@(installed_names)'
 4     *ast.PromiseGuard 'classes:'
 6       *ast.Promiser '"/usr/sbin/tcpdump"'
 8         *ast.Constraint 'perms'
10           *ast.FatArrow '=>'
10           *ast.Function 'mog'
12             *ast.GiveArgItem
14               *ast.Identifier '0555'
12             *ast.GiveArgItem
14               *ast.Identifier 'root'
12             *ast.GiveArgItem
14               *ast.Identifier 'root'
~~~

## Autofmt in (neo)vim

~~~
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin") " fmt
au BufWritePost *.cf silent call Fmt("cffmt /dev/stdin") " fmt on save
~~~

## Developing

You'll need goyacc, and then 'go generate', go build and then possibly also build cmd/cffmt. Files
of most interest are `parse.y` and `print.go`. The lexer (`lex.go`) is mostly doing the right thing.
