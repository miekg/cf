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
standard output. Cf uses an indent of 2 spaces to indent elements of the tree. Look for example at
the following snippet and the pretty printed on after it.

~~~ cfengine
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

If the first line of the file contains the comment: `# cffmt:no`  it will not be pretty printed.
Another directive is `# cffmt:list-nl` which says the _next_ list will have each item printed on a
new line. For example:

~~~ cfengine
"Clients"         or => { aaa, bbb, ccc, dddd, eee, fff,
  ggg, hhhh };
~~~

Normally to:

~~~ cfengine
"Clients"         or => { aaa, bbb, ccc, dddd, eee, fff, ggg, hhhh };
~~~

But with `# cffmt:list-nl`:

~~~ cfengine
# cffmt:list-nl
"Clients"         or => { aaa,
                          bbb,
                          ccc,
                          dddd,
                          eee,
                          fff,
                          ggg,
                          hhhh };
~~~

Options are:

`-a`
:   print the AST to standard error

`-p`
:   print the pretty printed document to standard output (defaults to true)

`-l`
:   show lexer tokens

`-d`
:   when parsing fails print the debug tree

## See Also

See the project's README for more details. Source code and development takes place on
[GitHub](https://github.com/miekg/cf).

## Author

Miek Gieben <miek@miek.nl>.
