package fileredactor

import (
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/go-utils/v2/filterwriter"
	"github.com/bitrise-io/go-utils/v2/log"
)

const bufferSize = 1024

type FileRedactor interface {
	RedactFiles([]string, []string) error
}

type fileRedactor struct {
}

func NewFileRedactor() FileRedactor {
	return fileRedactor{}
}

func (f fileRedactor) RedactFiles(filePaths []string, secrets []string) error {
	logger := log.NewLogger()
	for _, path := range filePaths {
		if err := f.redactFile(path, secrets, logger); err != nil {
			return fmt.Errorf("failed to redact file (%s): %w", path, err)
		}
	}

	return nil
}

func (f fileRedactor) redactFile(path string, secrets []string, logger log.Logger) error {
	source, err := os.Open(path)
	defer source.Close()

	newPath := path + ".redacted"
	destination, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for redaction: %w", err)
	}
	defer destination.Close()

	buffer := make([]byte, bufferSize)
	redactWriter := filterwriter.New(secrets, destination, logger)
	for {
		n, err := source.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := redactWriter.Write(buffer[:n]); err != nil {
			return err
		}
	}

	//rename new file to old file name
	err = os.Rename(path, newPath)
	if err != nil {
		return fmt.Errorf("failed to overwrite old file (%s) with redacted file: %w", path, err)
	}
	
	return nil
}
