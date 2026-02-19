// Package transport provides STDIO transport for MCP client communication.
package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/mcp"
)

// STDIOTransport handles communication with MCP servers via STDIO
type STDIOTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr io.Reader
	mu     sync.Mutex
	closed bool
}

// NewSTDIOTransport creates a new STDIO transport for the given command
func NewSTDIOTransport(command string, args []string, env []string) (*STDIOTransport, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	return &STDIOTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: stderr,
	}, nil
}

// Start starts the MCP server process
func (t *STDIOTransport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	logger.InfoC("mcp.transport", fmt.Sprintf("Starting MCP server: %s %v", t.cmd.Path, t.cmd.Args))

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Start goroutine to log stderr
	go t.logStderr()

	return nil
}

// logStderr logs stderr output from the MCP server
func (t *STDIOTransport) logStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		logger.DebugC("mcp.transport", fmt.Sprintf("MCP server stderr: %s", scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		logger.ErrorC("mcp.transport", fmt.Sprintf("Error reading stderr: %v", err))
	}
}

// Send sends a JSON-RPC request to the MCP server
func (t *STDIOTransport) Send(ctx context.Context, req *mcp.JSONRPCRequest) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.DebugC("mcp.transport", fmt.Sprintf("Sending: %s", string(data)))

	if _, err := t.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	if _, err := t.stdin.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// Receive receives a JSON-RPC response from the MCP server
func (t *STDIOTransport) Receive(ctx context.Context) (*mcp.JSONRPCResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	line, err := t.stdout.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("MCP server closed connection")
		}
		return nil, fmt.Errorf("failed to read from stdout: %w", err)
	}

	logger.DebugC("mcp.transport", fmt.Sprintf("Received: %s", line))

	var resp mcp.JSONRPCResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Close closes the transport and terminates the MCP server
func (t *STDIOTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	logger.InfoC("mcp.transport", "Closing MCP server transport")

	// Close stdin first to signal EOF to the server
	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			logger.ErrorC("mcp.transport", fmt.Sprintf("Error closing stdin: %v", err))
		}
	}

	// Wait for the process to exit (with timeout)
	if t.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- t.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				logger.WarnC("mcp.transport", fmt.Sprintf("MCP server exited with error: %v", err))
			} else {
				logger.InfoC("mcp.transport", "MCP server exited cleanly")
			}
		}
	}

	return nil
}

// IsClosed returns true if the transport is closed
func (t *STDIOTransport) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}
