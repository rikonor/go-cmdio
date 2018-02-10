package cmdio

import (
	"io"
	"log"
	"os"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// CloseFunc needs to be called to clean up resources
type CloseFunc func() error

// Wrap wraps a command such that it uses named pipes attached to the given reader
// and writer instead of using actual files.
func Wrap(r io.Reader, w io.Writer, execArgs []string) ([]string, CloseFunc, error) {
	// Setup pipe names
	inPipe := "inPipe"
	outPipe := "outPipe"

	// Prepare args
	for i := 0; i < len(execArgs); i++ {
		switch execArgs[i] {
		case "INPUT":
			execArgs[i] = inPipe
		case "OUTPUT":
			execArgs[i] = outPipe
		}
	}

	// Setup named pipes
	if err := syscall.Mkfifo(inPipe, 0644); err != nil {
		return nil, nil, err
	}

	if err := syscall.Mkfifo(outPipe, 0644); err != nil {
		return nil, nil, err
	}

	eg := &errgroup.Group{}

	// Pipe input
	eg.Go(func() error {
		fIn, err := os.OpenFile(inPipe, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}
		defer fIn.Close()

		_, err = io.Copy(fIn, r)
		return err
	})

	// Pipe output
	eg.Go(func() error {
		fOut, err := os.OpenFile(outPipe, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatal(err)
		}
		defer fOut.Close()

		_, err = io.Copy(w, fOut)
		return err
	})

	// Setup cleanup function
	closeFn := func() error {
		if err := os.Remove(inPipe); err != nil {
			return err
		}

		if err := os.Remove(outPipe); err != nil {
			return err
		}

		return eg.Wait()
	}

	return execArgs, closeFn, nil
}
