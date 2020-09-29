package repl

import (
	"bufio"
	"fmt"
	"github.com/mark07x/TLang/evaluator"
	"github.com/mark07x/TLang/lexer"
	"github.com/mark07x/TLang/object"
	"github.com/mark07x/TLang/parser"
	"io"
)

const PROMPT = "T> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment(evaluator.Bases)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			fmt.Printf("\n")
			scanner = bufio.NewScanner(in)
			continue
		}

		line := scanner.Text()
		for len(line) != 0 && line[len(line)-1] == '\\' {
			line = line[:len(line)-1]
			fmt.Printf(".. ")
			scanned := scanner.Scan()
			if !scanned {
				fmt.Printf("\n")
				scanner = bufio.NewScanner(in)
				continue
			}
			line = line + scanner.Text()
		}
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			evaluator.PrintParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != object.VoidObj {
			_, _ = io.WriteString(out, evaluated.Inspect())
			_, _ = io.WriteString(out, "\n")
		}
	}
}
