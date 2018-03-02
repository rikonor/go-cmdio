package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	cmdio "github.com/rikonor/go-cmdio"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: stdio-wrapper <executable> <args>")
		os.Exit(1)
	}

	// Prepare command
	execPath := os.Args[1]
	execArgs := os.Args[2:]

	// Wrap command
	r := os.Stdin
	w := os.Stdout

	tmpArgs, closeFn, err := cmdio.WrapSimple(r, w, execArgs)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFn()

	// Run command
	cmd := exec.Command(execPath, tmpArgs...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
