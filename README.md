# CFengine pretty printer

Experimental for now - but I believe I got *most* of the syntax elements right. Found a few examples
of CFEngine files that are now placed in 'unparseable'.

Comments are left alone and outputted as-is. This may change in the future and we might do something
that is similar to what Go is doing. Long lists are wrapped and indented slightly.

miekg/cf uses the lexer and parser from CFengine's source and converts it into a (Go) AST that we
can walk and print.

Install with: `go install github.com/miekg/cf/cmd/cffmt@latest`

Will not correctly parse:

- drops comments that are placed in a bundle/body but at the end.
- thinarrow is not parsed yet.

## Autofmt in (neo)vim

~~~
au Filetype cf3 normal zR
au FileType cf3 command! Fmt call Fmt("cffmt /dev/stdin")
~~~

## Usage

Build `cffmt` in the cmd/cffmt and then for an example:

    ./cffmt ../../testdata/promtest.cf

If you only want the AST use -a, and throw away standard output:

    ./cffmt -a /home/miek/src/github.com/miekg/playground/cfjson/cf/list.cf > /dev/nul

This lists:

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

Where the left side number is the number of spaces for the indentation (to easily identify if nodes
are on the same level).
