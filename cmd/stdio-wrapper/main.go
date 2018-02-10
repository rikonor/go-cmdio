package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: stdio-wrapper <executable> <args>")
		os.Exit(1)
	}

	inPipe := "inPipe"
	outPipe := "outPipe"

	// Prepare command
	execPath := os.Args[1]
	execArgs := os.Args[2:]

	for i := 0; i < len(execArgs); i++ {
		switch execArgs[i] {
		case "INPUT":
			execArgs[i] = inPipe
		case "OUTPUT":
			execArgs[i] = outPipe
		}
	}

	cmd := exec.Command(execPath, execArgs...)

	// Setup named pipes
	eg := &errgroup.Group{}

	if err := syscall.Mkfifo(inPipe, 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(inPipe)

	if err := syscall.Mkfifo(outPipe, 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(outPipe)

	eg.Go(func() error {
		fIn, err := os.OpenFile(inPipe, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}
		defer fIn.Close()

		_, err = io.Copy(fIn, os.Stdin)
		return err
	})

	eg.Go(func() error {
		fOut, err := os.OpenFile(outPipe, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}
		defer fOut.Close()

		_, err = io.Copy(os.Stdout, fOut)
		return err
	})

	// Run command
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
