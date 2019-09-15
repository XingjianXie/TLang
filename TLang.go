package main

import (
	"TLang/evaluator"
	"TLang/lexer"
	"TLang/object"
	"TLang/parser"
	"TLang/repl"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) == 2 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
		env := object.NewEnvironment()
		l := lexer.New(string(data))
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			evaluator.PrintParserErrors(os.Stderr, p.Errors())
			os.Exit(1)
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated.Type() == object.ERR {
			_, _ = io.WriteString(os.Stderr, evaluated.Inspect())
			_, _ = io.WriteString(os.Stderr, "\n")
			os.Exit(1)
		}
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Printf("Welcome to T language!\n")
		repl.Start(os.Stdin, os.Stdout)
	} else {
		fmt.Printf("Usage: " + os.Args[0] + " <file>")
	}
}
