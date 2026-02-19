// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// ToolWrapper adapts MCP tools to picoclaw's Tool interface
type ToolWrapper struct {
	client *Client
	tool   MCPTool
}

// NewToolWrapper creates a new tool wrapper for an MCP tool
func NewToolWrapper(client *Client, tool MCPTool) *ToolWrapper {
	return &ToolWrapper{
		client: client,
		tool:   tool,
	}
}

// Name returns the tool name with server prefix
func (w *ToolWrapper) Name() string {
	return fmt.Sprintf("mcp_%s_%s", w.client.GetName(), w.tool.Name)
}

// Description returns the tool description with server prefix
func (w *ToolWrapper) Description() string {
	return fmt.Sprintf("[MCP:%s] %s", w.client.GetName(), w.tool.Description)
}

// Parameters returns the tool's input schema
func (w *ToolWrapper) Parameters() map[string]interface{} {
	return w.tool.InputSchema
}

// Execute executes the MCP tool with the given arguments
func (w *ToolWrapper) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	result, err := w.client.CallTool(ctx, w.tool.Name, args)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("MCP tool %s error: %v", w.tool.Name, err))
	}
	return tools.NewToolResult(result)
}

// GetMCPTool returns the underlying MCP tool definition
func (w *ToolWrapper) GetMCPTool() MCPTool {
	return w.tool
}

// GetClient returns the underlying MCP client
func (w *ToolWrapper) GetClient() *Client {
	return w.client
}
