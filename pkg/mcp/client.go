// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/mcp/transport"
)

// Client represents an MCP client connected to a single server
type Client struct {
	server       MCPServer
	transport    *transport.STDIOTransport
	requestID    int64
	capabilities ServerCapabilities
	tools        []MCPTool
	initialized  bool
	mu           sync.Mutex
}

// NewClient creates a new MCP client
func NewClient(server MCPServer) (*Client, error) {
	stdioTransport, err := transport.NewSTDIOTransport(server.Command, server.Args, server.Env)
	if err != nil {
		return nil, fmt.Errorf("failed to create STDIO transport: %w", err)
	}

	return &Client{
		server:    server,
		transport: stdioTransport,
		requestID: 0,
		tools:     make([]MCPTool, 0),
	}, nil
}

// Connect connects to the MCP server
func (c *Client) Connect(ctx context.Context) error {
	logger.InfoC("mcp.client", fmt.Sprintf("Connecting to MCP server: %s", c.server.Name))

	if err := c.transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	return nil
}

// Initialize performs the initialize handshake with the MCP server
func (c *Client) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.InfoC("mcp.client", fmt.Sprintf("Initializing MCP server: %s", c.server.Name))

	c.requestID++

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.requestID,
		"method":  MethodInitialize,
		"params": InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    ClientCaps{},
			ClientInfo: ClientInfo{
				Name:    "picoclaw",
				Version: "0.1.0",
			},
		},
	}

	if err := c.transport.Send(ctx, req); err != nil {
		return fmt.Errorf("failed to send initialize: %w", err)
	}

	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive initialize response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	var result InitializeResult
	if err := unmarshalJSON(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	c.capabilities = result.Capabilities
	c.initialized = true

	// Send initialized notification
	notif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  MethodInitialized,
	}
	if err := c.transport.Send(ctx, notif); err != nil {
		logger.WarnC("mcp.client", fmt.Sprintf("Failed to send initialized notification: %v", err))
	}

	logger.InfoC("mcp.client", fmt.Sprintf("MCP server initialized: %s (version %s)",
		result.ServerInfo.Name, result.ServerInfo.Version))

	return nil
}

// ListTools retrieves the list of available tools from the MCP server
func (c *Client) ListTools(ctx context.Context) ([]MCPTool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	logger.DebugC("mcp.client", fmt.Sprintf("Listing tools from server: %s", c.server.Name))

	c.requestID++

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.requestID,
		"method":  MethodListTools,
		"params":  ListToolsParams{},
	}

	if err := c.transport.Send(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to send list_tools: %w", err)
	}

	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive list_tools response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list_tools error: %s", resp.Error.Message)
	}

	var result ListToolsResult
	if err := unmarshalJSON(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list_tools result: %w", err)
	}

	c.tools = result.Tools

	logger.InfoC("mcp.client", fmt.Sprintf("Server %s has %d tools", c.server.Name, len(result.Tools)))

	return result.Tools, nil
}

// CallTool calls a tool on the MCP server
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return "", fmt.Errorf("client not initialized")
	}

	logger.InfoCF("mcp.client", fmt.Sprintf("Calling tool %s on server %s", name, c.server.Name),
		map[string]interface{}{"args": args})

	c.requestID++

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.requestID,
		"method":  MethodCallTool,
		"params": CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}

	if err := c.transport.Send(ctx, req); err != nil {
		return "", fmt.Errorf("failed to send tools/call: %w", err)
	}

	resp, err := c.transport.Receive(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to receive tools/call response: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("tools/call error: %s", resp.Error.Message)
	}

	var result ToolCallResult
	if err := unmarshalJSON(resp.Result, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal tools/call result: %w", err)
	}

	text := result.GetText()
	logger.InfoC("mcp.client", fmt.Sprintf("Tool %s returned %d bytes", name, len(text)))

	return text, nil
}

// GetTools returns the cached list of tools
func (c *Client) GetTools() []MCPTool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tools
}

// GetName returns the server name
func (c *Client) GetName() string {
	return c.server.Name
}

// IsInitialized returns true if the client has been initialized
func (c *Client) IsInitialized() bool {
	return atomic.LoadInt64(&c.requestID) > 0
}

// Close closes the connection to the MCP server
func (c *Client) Close() error {
	logger.InfoC("mcp.client", fmt.Sprintf("Closing MCP client: %s", c.server.Name))

	c.mu.Lock()
	c.initialized = false
	c.mu.Unlock()

	return c.transport.Close()
}

// Shutdown sends a shutdown request to the MCP server
func (c *Client) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil
	}

	logger.InfoC("mcp.client", fmt.Sprintf("Shutting down MCP server: %s", c.server.Name))

	c.requestID++

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      c.requestID,
		"method":  MethodShutdown,
	}

	if err := c.transport.Send(ctx, req); err != nil {
		logger.WarnC("mcp.client", fmt.Sprintf("Failed to send shutdown: %v", err))
	}

	// Don't wait for response, just close
	return c.transport.Close()
}

// Helper function to unmarshal JSON
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
