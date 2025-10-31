package output

import (
	"bufio"
	"fmt"
	"os"
)

// FileOutput writes logs to a file.
type FileOutput struct {
	file   *os.File
	writer *bufio.Writer
	path   string
}

// NewFileOutput creates a new file output.
func NewFileOutput(config map[string]interface{}) (*FileOutput, error) {
	pathVal, ok := config["path"]
	if !ok {
		return nil, fmt.Errorf("file output requires 'path' in config")
	}

	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("file output 'path' must be a string")
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return &FileOutput{
		file:   file,
		writer: bufio.NewWriterSize(file, 1024*1024), // 1MB buffer
		path:   path,
	}, nil
}

// WriteBatch writes a batch of log lines to the file.
func (f *FileOutput) WriteBatch(logs []string) error {
	for _, log := range logs {
		if _, err := f.writer.WriteString(log); err != nil {
			return fmt.Errorf("failed to write log: %w", err)
		}
		if _, err := f.writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	// Flush to disk
	if err := f.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// Close closes the file.
func (f *FileOutput) Close() error {
	if err := f.writer.Flush(); err != nil {
		return err
	}
	return f.file.Close()
}

// Name returns the output name.
func (f *FileOutput) Name() string {
	return fmt.Sprintf("file:%s", f.path)
}
