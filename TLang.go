package main

import (
	"fmt"
	"github.com/mark07x/TLang/evaluator"
	"github.com/mark07x/TLang/lexer"
	"github.com/mark07x/TLang/object"
	"github.com/mark07x/TLang/parser"
	"github.com/mark07x/TLang/repl"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) == 2 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			print(err)
			os.Exit(1)
		}
		env := evaluator.SharedEnv.NewEnclosedEnvironment()
		l := lexer.New(string(data))
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			evaluator.PrintParserErrors(os.Stderr, p.Errors())
			os.Exit(1)
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated.Type() == object.ERR {
			_, _ = io.WriteString(os.Stderr, evaluated.Inspect(16))
			_, _ = io.WriteString(os.Stderr, "\n")
			os.Exit(1)
		}
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Printf("Welcome to T Language!\n")
		repl.Start(os.Stdin, os.Stdout)
	} else {
		fmt.Printf("Usage: " + os.Args[0] + " <file>")
	}
}
