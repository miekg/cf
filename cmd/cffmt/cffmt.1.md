%%%
title = "cffmt 1"
area = "User Commands"
workgroup = "CFEngine"
%%%

cffmt
=====

## Name

cffmt - CFEngine pretty printer/formatter

## Synopsis

cffmt  *[OPTION]*... *[FILE]*

## Description

Cffmt will parse a CFengine file in **FILE** or from standard input and will pretty print it to
standard output.

If the first line of the file contains the comment: `# cffmt:no`  it will not be pretty printed.
Another directive is `# cffmt:list-nl` which says the _next_ list will have each item printed on a
new line.

Options are:

`-a`
:   print the AST to standard error

`-p`
:   print the pretty printed document to standard output (defaults to true)

`-l`
:   show lexer tokens

`-f`
:   if parsing fails only show the name of the file being parsed.

## See Also

See the project's README for more details. Source code and development takes place on
[GitHub](https://github.com/miekg/cf).

## Author

Miek Gieben <miek@miek.nl>.
