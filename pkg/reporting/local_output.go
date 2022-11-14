package reporting

import (
	"fmt"
	"io"
	"os"
)

// writer is a writer for local report output. We're doing this so that we can avoid closing stdout
// if we're writing to that instead of an actual file.
type writer struct {
	file *os.File

	io.Writer
	io.Closer
}

// createWriter will create a writer that points to the given file described by output.
// If the string is "-", the writer will point to os.Stdout instead.
func createWriter(output string) (*writer, error) {
	writer := &writer{}
	if output == "-" {
		writer.file = os.Stdout
	} else {
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return nil, fmt.Errorf("error opening output file for writing: %v", err)
		}

		writer.file = file
	}

	return writer, nil
}

// Write calls the underlying writer implementation.
func (w *writer) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

// Close closes the writer if possible
func (w *writer) Close() error {
	if w.file == os.Stdout {
		return nil
	}

	return w.file.Close()
}
