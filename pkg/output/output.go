// Package output provides implementations for different log output destinations
package output

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/P1llus/genlog/pkg/config"
)

// Output represents a destination for log messages
type Output interface {
	// Write sends a batch of log messages to the destination
	Write(messages []string) error
	// Close closes the output and releases any resources
	Close() error
}

// LogGenerator represents the interface needed for generating log lines
type LogGenerator interface {
	GenerateLogLine() (string, error)
}

// NewOutput creates a new output based on the configuration
func NewOutput(cfg config.OutputConfig, workerID int) (Output, error) {
	switch cfg.Type {
	case config.OutputTypeFile:
		return newFileOutput(cfg, workerID)
	case config.OutputTypeUDP:
		return newUDPOutput(cfg)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", cfg.Type)
	}
}

// fileOutput implements Output for file destinations
type fileOutput struct {
	file     *os.File
	writer   *bufio.Writer
	filename string
	mu       sync.Mutex
}

func newFileOutput(cfg config.OutputConfig, workerID int) (*fileOutput, error) {
	filename, ok := cfg.Config["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename is required for file output")
	}

	// If there are multiple workers, append worker ID to filename
	if cfg.Workers > 1 {
		ext := filepath.Ext(filename)
		base := filename[:len(filename)-len(ext)]
		filename = fmt.Sprintf("%s_worker%d%s", base, workerID, ext)
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("error creating output file: %w", err)
	}

	return &fileOutput{
		file:     file,
		writer:   bufio.NewWriter(file),
		filename: filename,
	}, nil
}

func (o *fileOutput) Write(messages []string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, msg := range messages {
		if _, err := o.writer.WriteString(msg + "\n"); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}
	return o.writer.Flush()
}

func (o *fileOutput) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.writer.Flush(); err != nil {
		return err
	}
	return o.file.Close()
}

// udpOutput implements Output for UDP destinations
type udpOutput struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	writeBuf int
	mu       sync.Mutex // Protects conn during concurrent writes from same worker
}

func newUDPOutput(cfg config.OutputConfig) (*udpOutput, error) {
	addrStr, ok := cfg.Config["address"].(string)
	if !ok {
		return nil, fmt.Errorf("address is required for UDP output")
	}

	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %w", err)
	}

	writeBuf := 1024 * 1024 // Default 1MB buffer
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP connection: %w", err)
	}

	// Set write buffer size for better performance
	if err := conn.SetWriteBuffer(writeBuf); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error setting write buffer: %w", err)
	}

	return &udpOutput{
		conn:     conn,
		addr:     addr,
		writeBuf: writeBuf,
	}, nil
}

func (o *udpOutput) Write(messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	// Lock to prevent concurrent writes to the same connection
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, msg := range messages {
		// Add newline to message
		msgBytes := []byte(msg + "\n")

		// Check if message exceeds UDP packet size
		if len(msgBytes) > 1472 { // Max safe UDP payload size
			// Truncate large messages
			msgBytes = msgBytes[:1472]
		}

		if _, err := o.conn.Write(msgBytes); err != nil {
			return fmt.Errorf("error writing UDP message: %w", err)
		}
	}

	return nil
}

func (o *udpOutput) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.conn.Close()
}

// Worker represents a worker that generates and sends logs
type Worker struct {
	Output    Output
	generator LogGenerator
	batchSize int
	maxCount  int
	stopChan  chan struct{}
}

// NewWorker creates a new worker instance
func NewWorker(output Output, gen LogGenerator, batchSize, maxCount int, stopChan chan struct{}) *Worker {
	return &Worker{
		Output:    output,
		generator: gen,
		batchSize: batchSize,
		maxCount:  maxCount,
		stopChan:  stopChan,
	}
}

// Start begins the worker's log generation and sending process
func (w *Worker) Start() {
	batch := make([]string, 0, w.batchSize)
	ticker := time.NewTicker(100 * time.Millisecond) // Adjust batch timing as needed
	count := 0

	for {
		select {
		case <-w.stopChan:
			// Flush any remaining logs
			if len(batch) > 0 {
				if err := w.Output.Write(batch); err != nil {
					fmt.Printf("Error writing final batch: %v\n", err)
				}
			}
			return
		case <-ticker.C:
			if len(batch) >= w.batchSize {
				if err := w.Output.Write(batch); err != nil {
					fmt.Printf("Error writing batch: %v\n", err)
				}
				batch = batch[:0]
			}
		default:
			if w.maxCount > 0 && count >= w.maxCount {
				// We've reached our count limit
				if len(batch) > 0 {
					if err := w.Output.Write(batch); err != nil {
						fmt.Printf("Error writing final batch: %v\n", err)
					}
				}
				return
			}

			logLine, err := w.generator.GenerateLogLine()
			if err != nil {
				fmt.Printf("Error generating log line: %v\n", err)
				continue
			}
			batch = append(batch, logLine)
			count++
		}
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.stopChan)
}
