package output

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type writer interface {
	write(name string, data []byte) error
}

type dirWriter struct {
	basePath string
}

type fileWriter struct {
	filePath    string
	initialized bool
}

type stdoutWriter struct {
	writer io.Writer
}

func (fw *fileWriter) write(name string, data []byte) error {
	var file *os.File

	var err error

	if !fw.initialized {
		if err := os.MkdirAll(filepath.Dir(fw.filePath), 0700); err != nil {
			return fmt.Errorf("failed to create parent dir for output file: %w", err)
		}

		file, err = os.Create(fw.filePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}

		fw.initialized = true
	} else {
		file, err = os.OpenFile(fw.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
	}

	defer func() { _ = file.Close() }()

	if _, err := file.Write(append(outputHeader(name), data...)); err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	return nil
}

func (dw dirWriter) write(name string, data []byte) error {
	if err := os.MkdirAll(dw.basePath, 0700); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", dw.basePath, err)
	}

	path := path.Join(dw.basePath, name)

	return os.WriteFile(path, data, 0600)
}

func (sw stdoutWriter) write(name string, data []byte) error {
	if sw.writer == nil {
		sw.writer = os.Stdout
	}

	if _, err := sw.writer.Write(append(outputHeader(name), data...)); err != nil {
		return err
	}

	return nil
}

func outputHeader(name string) []byte {
	return []byte(fmt.Sprintf("---\n# %s\n", name))
}
