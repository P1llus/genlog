package output

import (
	"fmt"
	"net"
)

// UDPOutput sends logs via UDP.
type UDPOutput struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

// NewUDPOutput creates a new UDP output.
func NewUDPOutput(config map[string]interface{}) (*UDPOutput, error) {
	addressVal, ok := config["address"]
	if !ok {
		return nil, fmt.Errorf("udp output requires 'address' in config")
	}

	address, ok := addressVal.(string)
	if !ok {
		return nil, fmt.Errorf("udp output 'address' must be a string")
	}

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address %s: %w", address, err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to UDP address %s: %w", address, err)
	}

	// Set write buffer size for better performance
	if err := conn.SetWriteBuffer(1024 * 1024); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set write buffer: %w", err)
	}

	return &UDPOutput{
		conn: conn,
		addr: addr,
	}, nil
}

// WriteBatch writes a batch of log lines via UDP.
func (u *UDPOutput) WriteBatch(logs []string) error {
	for _, log := range logs {
		// Add newline
		data := []byte(log + "\n")

		// UDP has a max packet size, truncate if needed
		if len(data) > 1472 { // Safe UDP payload size
			data = data[:1472]
		}

		if _, err := u.conn.Write(data); err != nil {
			return fmt.Errorf("failed to write UDP packet: %w", err)
		}
	}

	return nil
}

// Close closes the UDP connection.
func (u *UDPOutput) Close() error {
	return u.conn.Close()
}

// Name returns the output name.
func (u *UDPOutput) Name() string {
	return fmt.Sprintf("udp:%s", u.addr.String())
}
