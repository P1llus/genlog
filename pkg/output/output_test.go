package output

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/P1llus/genlog/pkg/config"
)

// mockGenerator implements LogGenerator for testing
type mockGenerator struct {
	lines []string
	index int
}

func (m *mockGenerator) GenerateLogLine() (string, error) {
	if m.index >= len(m.lines) {
		m.index = 0
	}
	line := m.lines[m.index]
	m.index++
	return line, nil
}

func TestFileOutput(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create output config
	cfg := config.OutputConfig{
		Type:    config.OutputTypeFile,
		Workers: 1,
		Config: map[string]interface{}{
			"filename": filepath.Join(tmpDir, "test.log"),
		},
	}

	// Create output
	out, err := NewOutput(cfg, 0)
	if err != nil {
		t.Fatalf("NewOutput failed: %v", err)
	}
	defer out.Close()

	// Test writing messages
	messages := []string{
		"test message 1",
		"test message 2",
		"test message 3",
	}

	if err := out.Write(messages); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back the messages
	file, err := os.Open(cfg.Config["filename"].(string))
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i, msg := range messages {
		if !scanner.Scan() {
			t.Fatalf("Expected to read message %d", i)
		}
		if scanner.Text() != msg {
			t.Errorf("Message %d mismatch: got %s, want %s", i, scanner.Text(), msg)
		}
	}
}

func TestUDPOutput(t *testing.T) {
	// Start a UDP server to receive messages
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Create output config
	cfg := config.OutputConfig{
		Type:    config.OutputTypeUDP,
		Workers: 1,
		Config: map[string]interface{}{
			"address": conn.LocalAddr().String(),
		},
	}

	// Create output
	out, err := NewOutput(cfg, 0)
	if err != nil {
		t.Fatalf("NewOutput failed: %v", err)
	}
	defer out.Close()

	// Test writing messages
	messages := []string{
		"test message 1",
		"test message 2",
		"test message 3",
	}

	if err := out.Write(messages); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	buf := make([]byte, 1024)
	for i, msg := range messages {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			t.Fatalf("Failed to read message %d: %v", i, err)
		}
		received := string(buf[:n-1]) // Remove trailing newline
		if received != msg {
			t.Errorf("Message %d mismatch: got %s, want %s", i, received, msg)
		}
	}
}

func TestWorker(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "worker-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create output config
	cfg := config.OutputConfig{
		Type:    config.OutputTypeFile,
		Workers: 1,
		Config: map[string]interface{}{
			"filename": filepath.Join(tmpDir, "test.log"),
		},
	}

	// Create output
	out, err := NewOutput(cfg, 0)
	if err != nil {
		t.Fatalf("NewOutput failed: %v", err)
	}
	defer out.Close()

	// Create mock generator
	gen := &mockGenerator{
		lines: []string{
			"test message 1",
			"test message 2",
			"test message 3",
		},
	}

	// Create worker
	stopChan := make(chan struct{})
	worker := NewWorker(out, gen, 2, 3, stopChan)

	// Start worker
	go worker.Start()

	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	close(stopChan)

	// Read back the messages
	file, err := os.Open(cfg.Config["filename"].(string))
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 messages, got %d", count)
	}
}

func TestInvalidOutputType(t *testing.T) {
	cfg := config.OutputConfig{
		Type:    "invalid",
		Workers: 1,
		Config: map[string]interface{}{
			"filename": "test.log",
		},
	}

	_, err := NewOutput(cfg, 0)
	if err == nil {
		t.Error("Expected error for invalid output type, got nil")
	}
}

func TestMissingConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.OutputConfig
	}{
		{
			name: "missing filename",
			cfg: config.OutputConfig{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config:  map[string]interface{}{},
			},
		},
		{
			name: "missing address",
			cfg: config.OutputConfig{
				Type:    config.OutputTypeUDP,
				Workers: 1,
				Config:  map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOutput(tt.cfg, 0)
			if err == nil {
				t.Error("Expected error for missing config, got nil")
			}
		})
	}
}
