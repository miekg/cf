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
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("Expect cf file")
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spec := cf.Parse(f)

	if *flagAst {
		ast.Print(os.Stderr, spec)
	}
	if *flagPrint {
		doc := &bytes.Buffer{}
		cf.Print(doc, spec)
		fmt.Print(doc.String())
	}
}
