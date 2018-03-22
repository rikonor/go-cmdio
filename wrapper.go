package cmdio

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// CloseFunc needs to be called to clean up resources
type CloseFunc func() error

// Wrap wraps a command such that it uses named pipes attached to the given readers and writers
// instead of using actual files.
func Wrap(rs []io.Reader, ws []io.Writer, execArgs []string) ([]string, CloseFunc, error) {
	// Keep track of files to remove later
	toRemove := []string{}

	i := 0
	getNextName := func() string {
		i++
		return fmt.Sprintf("pipe%d", i)
	}

	eg := &errgroup.Group{}

	// Setup a slice of tmp args
	tmpArgs := make([]string, len(execArgs))

	for i := 0; i < len(execArgs); i++ {
		arg := execArgs[i]

		switch {
		// Handle input file
		case strings.HasPrefix(arg, "INPUT"):
			// Get reader index
			idxStr := strings.TrimPrefix(arg, "INPUT")
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to parse reader index from arg name")
			}

			pipeName := getNextName()
			tmpArgs[i] = pipeName

			// Create pipe file
			if err := syscall.Mkfifo(pipeName, 0644); err != nil {
				return nil, nil, errors.Wrap(err, "failed to create pipe")
			}

			// Add pipe to removal list
			toRemove = append(toRemove, pipeName)

			// Start copying from reader to pipe
			eg.Go(func() error {
				fIn, err := os.OpenFile(pipeName, os.O_WRONLY, os.ModeNamedPipe)
				if err != nil {
					return errors.Wrap(err, "failed to open pipe")
				}
				defer fIn.Close()

				_, err = io.Copy(fIn, rs[idx-1])
				return errors.Wrap(err, "failed to copy from reader to pipe")
			})

		// Handle output file
		case strings.HasPrefix(arg, "OUTPUT"):
			// Get writer index
			idxStr := strings.TrimPrefix(arg, "OUTPUT")
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to parse writer index from arg name")
			}

			pipeName := getNextName()
			tmpArgs[i] = pipeName

			// Create pipe file
			if err := syscall.Mkfifo(pipeName, 0644); err != nil {
				return nil, nil, errors.Wrap(err, "failed to create pipe")
			}

			// Add pipe to removal list
			toRemove = append(toRemove, pipeName)

			// Start copying from pipe to writer
			eg.Go(func() error {
				fOut, err := os.OpenFile(pipeName, os.O_RDONLY, os.ModeNamedPipe)
				if err != nil {
					return errors.Wrap(err, "failed to open pipe")
				}
				defer fOut.Close()

				_, err = io.Copy(ws[idx-1], fOut)
				return errors.Wrap(err, "failed to copy from pipe to writer")
			})

		// Use arg as-is
		default:
			tmpArgs[i] = arg
		}
	}

	// Setup cleanup function
	closeFn := func() error {
		// Remove pipe files
		for _, fPath := range toRemove {
			if err := os.Remove(fPath); err != nil {
				return errors.Wrap(err, "failed to remove pipe")
			}
		}

		return eg.Wait()
	}

	return tmpArgs, closeFn, nil
}

// WrapSimple wraps a command such that it uses named pipes attached to the given reader
// and writer instead of using actual files.
func WrapSimple(r io.Reader, w io.Writer, execArgs []string) ([]string, CloseFunc, error) {
	// Setup pipe names
	inPipe := "inPipe"
	outPipe := "outPipe"

	// Setup a slice of tmp args
	tmpArgs := make([]string, len(execArgs))

	// Prepare args
	for i := 0; i < len(execArgs); i++ {
		switch execArgs[i] {
		case "INPUT":
			tmpArgs[i] = inPipe
		case "OUTPUT":
			tmpArgs[i] = outPipe
		default:
			tmpArgs[i] = execArgs[i]
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

	return tmpArgs, closeFn, nil
}
