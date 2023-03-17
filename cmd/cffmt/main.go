package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/miekg/cf"
	"github.com/miekg/cf/token"
)

var (
	flagAst   = flag.Bool("a", false, "print AST to standard output when successfully parsed")
	flagPrint = flag.Bool("p", true, "pretty print the file")
	flagFail  = flag.Bool("f", false, "when failing to parse only print the filename")
	flagLex   = flag.Bool("l", false, "only show the tokens")
	flagDebug = flag.Bool("d", false, "when failing to parse print debug output")
)

func main() {
	flag.Parse()
	var (
		err    error
		buffer []byte
	)

	switch flag.NArg() {
	case 0:
		buffer, err = io.ReadAll(os.Stdin)
	case 1:
		buffer, err = os.ReadFile(flag.Arg(0))
	default:
		log.Fatal("Too many arguments")
	}
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := cf.Lex(string(buffer))
	if err != nil {
		log.Fatal(err)
	}
	if len(tokens) > 0 {
		if ct, ok := tokens[0].(chroma.Token); ok {
			if ct.Type == token.Comment && strings.HasPrefix(ct.Value, "# cffmt:no") {
				fmt.Printf("%s", buffer)
				return
			}
		}
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
		if *flagDebug {
			for _, token := range tokens {
				log.Printf("%T %v", token, token)
			}
			log.Print("Debug Tree:\n", debugTree)
			log.Print("Parse Tree:\n", parseTree)
		}
		log.Fatalf("Failed to parse: %s", err)
	}
	if *flagAst {
		log.Print("Parse Tree:\n", parseTree)
	}
	if *flagPrint {
		cf.Print(os.Stdout, parseTree)
	}
}
