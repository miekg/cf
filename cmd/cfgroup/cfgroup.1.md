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
functionals.cf and  promises.cf. If *GROUP*s are given on the command line the members
of those groups are printed.

For finding those files the following algorithm is used:

* if files are piped into cgroup's standard input, that data is used.
* if the cwd is inside a git repository and the basename is called 'cfengine' it will use the files from
    the current git repository.
* if the cwd is not in a cfengine git repository **cfgroup** will try /var/cfengine.

Options are:

`-l`
:   print all groups to standard output.

`-r` *HOST*
:   reverse lookup, show the classes for this specific host.

`-x` *GROUP*
:   list the hostnames from all specified *GROUP*s that are _not also_ in this specific *GROUP*. Mostly
    used to filter out "IsInactive" hosts: `cfgroup -x IsInactive IsWebserver` as an example.

`-o`
:   list the hostnames that only appear _once_ in the specified *GROUP*s.

`-n`
:   list the hostnames that only appear _more than once_ in the specified *GROUP*s, i.e. the
    opposite of `-o`.

`-d`
:   enable debug logging.

`-v`
:   show version.

## Examples

Feed cfgroup multiple files on standard input, and list the defined classes:

    cat /home/miek/src/gitlab.cncz.nl/sys/cfengine/masterfiles/adm/{groups.cf,functionals.cf} | ./cfgroup -l

## See Also

See the project's README for more details. Development takes place on [GitHub](https://github.com/miekg/cf).

## Author

Miek Gieben <miek@miek.nl>.
