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

cffmt` *[OPTION]*... *[FILE]*

## Description

Cffmt will take a CFengine file out of **FILE** or from standard input and will reformat it.
If the first line of the file contains the comment: `# cffmt:no`  it will not be formatted.

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

See the project's README for more details.

## Author

Miek Gieben <miek@miek.nl>.
