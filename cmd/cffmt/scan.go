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
	flagWidth = flag.Uint("w", 100, "width to use for list wrapping")
)

func main() {
	flag.Parse()
	l := &cf.Lexer{}
	switch flag.NArg() {
	case 0:
		l = cf.NewLexer(os.Stdin)
	case 1:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		l = cf.NewLexer(f, flag.Arg(0))
	default:
		log.Fatal("Too many arguments")
	}

	l.D = *flagDebug
	spec, err := cf.Parse(l)
	if err != nil {
		log.Fatalf("%s", err)
	}

	if *flagAst {
		ast.Print(os.Stderr, spec)
	}
	if *flagPrint {
		doc := &bytes.Buffer{}
		cf.PrintWithWidth(doc, *flagWidth, spec)
		fmt.Print(doc.String())
	}
}
