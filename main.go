package main

import (
	"TProject/evaluator"
	"TProject/lexer"
	"TProject/object"
	"TProject/parser"
	"TProject/repl"
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
			repl.PrintParserErrors(os.Stderr, p.Errors())
			os.Exit(1)
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			if evaluated.Type() == object.ERR {
				_, _ = io.WriteString(os.Stderr, evaluated.Inspect())
				_, _ = io.WriteString(os.Stderr, "\n")
				os.Exit(1)
			}
			_, _ = io.WriteString(os.Stdout, evaluated.Inspect())
			_, _ = io.WriteString(os.Stdout, "\n")
		}
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Printf("Welcome to T language!\n")
		repl.Start(os.Stdin, os.Stdout)
	} else {
		fmt.Printf("Usage: " + os.Args[0] + " <file>")
	}
}
