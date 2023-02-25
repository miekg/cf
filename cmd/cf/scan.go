package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/miekg/cf/ast"
)

var flagDebug = flag.Bool("d", false, "enable debugging")
var flagPrint = flag.Bool("p", true, "pretty print the file to standard output")
var flagAst = flag.Bool("a", false, "print AST")

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
	spec := ParseCF3(f)
	if *flagDebug {
		fmt.Println("****")
	}
	if *flagAst {
		ast.Print(os.Stdout, spec)
	}
	if *flagPrint {
		doc := &bytes.Buffer{}
		Print(doc, spec)
		fmt.Print(doc.String())
	}
}

func ParseCF3(r io.Reader) ast.Node {
	l := NewLexer(r, *flagDebug)
	yyParse(l)
	return l.Spec
}
