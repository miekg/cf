# CFengine pretty printer

Experimental for now - but I believe I got *most* of the syntax elements right.

This extracts the lexer and parser from CFengine's source and converts it into a (Go) AST that we
can walk and print.

## Usage

Build `cffmt` in the cmd/cffmt and then for an example:

    ./cffmt ../../testdata/promtest.cf
