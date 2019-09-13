package main

import (
	"TProject/repl"
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Welcome to T language!\n")
	repl.Start(os.Stdin, os.Stdout)
}
