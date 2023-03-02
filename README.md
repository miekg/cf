# CFEngine pretty printer

'cf' can handle most CFEngine files, a few files I found that aren't parseable are stored in the
'unparseable' directory.

'cf' will align fat-arrows in a constraint. And long lists are wrapped. If there is only 1
constraint it is printed on the same line and the promisers are aligned instead, an exception is
made for constraint that have 'contain => ...' or 'comment => ....'. Those are considered important
enough to be put "on the left".

Trailing commas of lists are removed.

Package cf uses the lexer and parser from CFEngine's source and converts it into a (Go) AST that we
can walk and print.

Install with: `go install github.com/miekg/cf/cmd/cffmt@latest`

Will not correctly parse:

- Comments that are placed in a bundle/body but at the end. These will be dropped.
- Multiline comments with escaped quoting characters.
- Will probably not work with Windows line endings.
- Macros are not parsed at all.

## TODO

- Thinarrow is not parsed yet. And possibly others syntax elements.
- Add tests with malformed content.
- promise guards don't have classguards as children, and they should.

## Usage

Build `cffmt` in the cmd/cffmt and then for an example:

    ./cffmt ../../testdata/promtest.cf

If you only want the AST use -a, and throw away standard output:

    ./cffmt -a /home/miek/src/github.com/miekg/playground/cfjson/cf/list.cf > /dev/nul

This shows the following. The left side number is the number of spaces for the indentation (to
easily identify if nodes are on the same level).

~~~
 0 *ast.Specification
 2   *ast.Bundle 'bundle'
 4     *ast.Identifier 'agent'
 4     *ast.Identifier 'one'
 4     *ast.PromiseGuard 'reports:'
 6       *ast.Promiser '"is_var"'
 8         *ast.Constraint 'if'
10           *ast.FatArrow '=>'
10           *ast.Function
12             *ast.Identifier 'isvariable'
12             *ast.GiveArgItem '"five"'
 6       *ast.Promiser '"two"'
 8         *ast.Constraint 'depends_on'
10           *ast.FatArrow '=>'
10           *ast.List
12             *ast.ListItem '"handle_one"'
12             *ast.ListItem '"handle_two"'
 6       *ast.Promiser '"one"'
 8         *ast.Constraint 'handle'
10           *ast.FatArrow '=>'
10           *ast.Qstring '"handle_one"'
 6       *ast.Promiser '"three"'
 8         *ast.Constraint 'handle'
10           *ast.FatArrow '=>'
10           *ast.Qstring '"10.5"'
~~~

## Autofmt in (neo)vim

~~~
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin") " fmt
au BufWritePost *.cf silent call Fmt("cffmt /dev/stdin") " fmt on save
~~~

## Developing

You'll need goyacc, and then 'go generate', go build and then possibly also build cmd/cffmt.
