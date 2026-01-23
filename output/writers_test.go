package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStdoutWriter(t *testing.T) {
	sw := stdoutWriter{}

	t.Run("write to stdout", func(t *testing.T) {
		err := sw.write("test-name", []byte("test-data"))
		if err != nil {
			t.Fatalf("expected no error writing to stdout, got: %v", err)
		}
	})

	t.Run("write to custom writer fails", func(t *testing.T) {
		sw := stdoutWriter{writer: &errorWriter{}}
		err := sw.write("test-name", []byte("test-data"))

		if err == nil || !strings.Contains(err.Error(), "failed") {
			t.Fatalf("expected \"failed\" error writing to failing writer, got: %v", err)
		}
	})
}

type errorWriter struct{}

func (e *errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("failed")
}

func TestDirWriter(t *testing.T) {
	dir := t.TempDir()
	fw := dirWriter{basePath: dir}
	err := fw.write("test-file-writer.txt", []byte("file-writer-test-data"))

	if err != nil {
		t.Fatalf("expected no error writing to file, got: %v", err)
	}

	// nolint: gosec
	data, err := os.ReadFile(filepath.Join(dir, "/test-file-writer.txt"))
	if err != nil {
		t.Fatalf("expected no error reading file, got: %v", err)
	}

	expected := "file-writer-test-data"
	if string(data) != expected {
		t.Fatalf("expected file contents to be %q, got %q", expected, string(data))
	}
}

func TestFileWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")
	fw := fileWriter{filePath: path}
	err := fw.write("part-1", []byte("file-writer-test-data-1"))

	if err != nil {
		t.Fatalf("expected no error writing to file, got: %v", err)
	}

	err = fw.write("part-2", []byte("file-writer-test-data-2"))

	if err != nil {
		t.Fatalf("expected no error writing to file, got: %v", err)
	}

	// nolint: gosec
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected no error reading file, got: %v", err)
	}

	expectSubStrings := []string{
		"part-1",
		"part-1",
		"file-writer-test-data-1",
		"file-writer-test-data-2",
	}

	for _, subString := range expectSubStrings {
		if !strings.Contains(string(got), subString) {
			t.Fatalf("expected file contents to contain %q, got %q", subString, string(got))
		}
	}
}
