package repl

import (
	"TLang/evaluator"
	"TLang/lexer"
	"TLang/object"
	"TLang/parser"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = "T> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		for len(line) != 0 && line[len(line)-1] == '\\' {
			line = line[:len(line)-1]
			fmt.Printf(".. ")
			scanned := scanner.Scan()
			if !scanned {
				return
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
		if evaluated != evaluator.VOID {
			_, _ = io.WriteString(out, evaluated.Inspect())
			_, _ = io.WriteString(out, "\n")
		}
	}
}
