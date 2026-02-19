// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/tools"
)

// Manager manages multiple MCP server connections
type Manager struct {
	clients  map[string]*Client
	registry *tools.ToolRegistry
	cfg      *config.Config
	mu       sync.RWMutex
}

// NewManager creates a new MCP manager
func NewManager(cfg *config.Config, toolRegistry *tools.ToolRegistry) *Manager {
	return &Manager{
		clients:  make(map[string]*Client),
		registry: toolRegistry,
		cfg:      cfg,
	}
}

// Start connects to all enabled MCP servers
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.InfoC("mcp.manager", "Starting MCP manager")

	for _, serverCfg := range m.cfg.MCP.Servers {
		if !serverCfg.Enabled {
			logger.DebugC("mcp.manager", fmt.Sprintf("Skipping disabled MCP server: %s", serverCfg.Name))
			continue
		}

		client, err := m.connectServer(ctx, serverCfg)
		if err != nil {
			logger.ErrorC("mcp.manager", fmt.Sprintf("Failed to connect to MCP server %s: %v", serverCfg.Name, err))
			continue
		}

		m.clients[serverCfg.Name] = client

		// Register MCP tools with picoclaw's ToolRegistry
		m.registerTools(client)
	}

	logger.InfoC("mcp.manager", fmt.Sprintf("MCP manager started with %d servers", len(m.clients)))

	return nil
}

// connectServer connects to a single MCP server
func (m *Manager) connectServer(ctx context.Context, serverCfg config.MCPServerConfig) (*Client, error) {
	server := MCPServer{
		Name:    serverCfg.Name,
		Command: serverCfg.Command,
		Args:    serverCfg.Args,
		Env:     serverCfg.Env,
		Enabled: serverCfg.Enabled,
	}

	logger.InfoC("mcp.manager", fmt.Sprintf("Connecting to MCP server: %s (%s %v)",
		serverCfg.Name, serverCfg.Command, serverCfg.Args))

	client, err := NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err := client.Initialize(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	// List tools
	_, err = client.ListTools(ctx)
	if err != nil {
		logger.WarnC("mcp.manager", fmt.Sprintf("Failed to list tools from %s: %v", serverCfg.Name, err))
	}

	logger.InfoC("mcp.manager", fmt.Sprintf("Connected to MCP server: %s", serverCfg.Name))

	return client, nil
}

// registerTools adapts MCP tools and registers them
func (m *Manager) registerTools(client *Client) error {
	mcpTools := client.GetTools()
	for _, tool := range mcpTools {
		wrapper := NewToolWrapper(client, tool)
		m.registry.Register(wrapper)
		logger.InfoC("mcp.manager", fmt.Sprintf("Registered MCP tool: %s", wrapper.Name()))
	}
	return nil
}

// Stop disconnects all MCP servers
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.InfoC("mcp.manager", "Stopping MCP manager")

	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			logger.ErrorC("mcp.manager", fmt.Sprintf("Error disconnecting MCP server %s: %v", name, err))
		}
	}

	m.clients = make(map[string]*Client)
}

// GetClient returns a client by name
func (m *Manager) GetClient(name string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[name]
	return client, ok
}

// ListClients returns all connected client names
func (m *Manager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}
