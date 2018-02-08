package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func main() {
	eg := &errgroup.Group{}

	inPipe := "inPipe"
	outPipe := "outPipe"

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

	buf := &bytes.Buffer{}

	eg.Go(func() error {
		fOut, err := os.OpenFile(outPipe, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}
		defer fOut.Close()

		_, err = io.Copy(buf, fOut)
		return err
	})

	cmd := exec.Command("./text-doubler", "inPipe", "outPipe")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, buf); err != nil {
		log.Fatal(err)
	}
}
