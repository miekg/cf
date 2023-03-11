package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/miekg/cf"
)

var (
	flagAst   = flag.Bool("a", false, "print AST to standard output when successfully parsed")
	flagPrint = flag.Bool("p", true, "pretty print the file")
	flagFail  = flag.Bool("f", false, "when failing to parse only print the filename")
	flagLex   = flag.Bool("l", false, "only show the tokens")
)

func main() {
	flag.Parse()
	var err error
	buffer := []byte{}

	switch flag.NArg() {
	case 0:
		buffer, err = io.ReadAll(os.Stdin)
	case 1:
		buffer, err = ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Too many arguments")
	}

	tokens, err := cf.Lex(string(buffer))
	if err != nil {
		log.Fatal(err)
	}
	if *flagLex {
		for _, token := range tokens {
			log.Printf("%T %v", token, token)
		}
		return
	}
	parseTree, debugTree, err := cf.ParseTokens(tokens)
	if parseTree == nil && debugTree == nil && err == nil {
		return
	}
	if err != nil {
		if *flagFail {
			log.Fatal(flag.Arg(0))
		}
		for _, token := range tokens {
			log.Printf("%T %v", token, token)
		}
		log.Print("Debug Tree:\n", debugTree)
		log.Print("Parse Tree:\n", parseTree)
		log.Fatalf("Failed to parse: %s", err)
	}
	if *flagAst {
		log.Print("Parse Tree:\n", parseTree)
	}
	if *flagPrint {
		cf.Print(os.Stdout, parseTree)
	}
}
