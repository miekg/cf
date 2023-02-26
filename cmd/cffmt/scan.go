package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miekg/cf"
	"github.com/miekg/cf/ast"
)

var (
	flagPrint = flag.Bool("p", true, "pretty print the file to standard output")
	flagAst   = flag.Bool("a", false, "print AST to standard error")
	flagDebug = flag.Bool("d", false, "enable debugging in the lexer and yacc")
)

func main() {
	flag.Parse()
	f := os.Stdin
	switch flag.NArg() {
	case 0:
	case 1:
		f1, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f1.Close()
		f = f1
	default:
		log.Fatal("Too many arguments")
	}

	l := cf.NewLexer(f)
	l.D = *flagDebug
	spec, err := cf.Parse(l)
	if err != nil {
		log.Fatalf("Error while parsing: %s", err)
	}

	if *flagAst {
		ast.Print(os.Stderr, spec)
	}
	if *flagPrint {
		doc := &bytes.Buffer{}
		cf.Print(doc, spec)
		fmt.Print(doc.String())
	}
}
