%%%
title = "cfgroup 1"
area = "User Commands"
workgroup = "CFEngine"
%%%

cfgroup
=====

## Name

cfgroup - list or query CFEngine configuration

## Synopsis

cfgroup  *[OPTION]*... *[GROUP]*...

## Description

Cfgroup will list or query CFengine configuration details. Note this is an internal tool that may
or may not be of use to you.

Cfgroup lists hostnames that are part of specified groups. Data sources are group.cf,
functionals.cf, promises.cf and schedule.cf. If *GROUP*s are given on the command line the members
of those groups are printed.

For find those files the following algorithm is used:

* if the cwd is inside a git repository and the root is called 'cfengine' it will use the files from
    the current git repository.
* if the cwd is not in a cfegine git repository **cfgroup** will try /var/cfengine.

Options are:

`-i` *FILE*[,*FILE*]...

A single *FILE* or a comma seperated list of *FILE*,*FILE* that should be used as input, typically
used for testing, but also useful to force cfgroup to parse a specific set of files.

`-l`
:   print all groups to standard output.

`-r` *HOST*
:   reverse lookup, show the classes for this specific host.

## See Also

See the project's README for more details. Development takes place on [GitHub](https://github.com/miekg/cf).

## Author

Miek Gieben <miek@miek.nl>.
