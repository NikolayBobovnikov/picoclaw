// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"encoding/json"
	"strings"
)

// MCPServer represents a configured MCP server
type MCPServer struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
	Enabled bool     `json:"enabled"`
}

// MCPTool represents a tool from MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ServerCapabilities from MCP handshake
type ServerCapabilities struct {
	Tools struct {
		ListChanged bool `json:"listChanged"`
	} `json:"tools"`
	Resources struct {
		Subscribe   bool `json:"subscribe"`
		ListChanged bool `json:"listChanged"`
	} `json:"resources"`
	Prompts struct {
		ListChanged bool `json:"listChanged"`
	} `json:"prompts"`
}

// ServerInfo from MCP initialize response
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientInfo for MCP initialize request
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the result of initialize call
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
}

// ToolCallResult is the result of a tool call
type ToolCallResult struct {
	Content []interface{} `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// TextContent represents text content in tool results
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ImageContent represents image content in tool results
type ImageContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text string `json:"text,omitempty"`
	Blob []byte `json:"blob,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling for ToolCallResult
func (t *ToolCallResult) UnmarshalJSON(data []byte) error {
	// Handle string response
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		t.Content = []interface{}{TextContent{Type: "text", Text: str}}
		return nil
	}

	// Handle array response
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		for _, item := range arr {
			var text TextContent
			if err := json.Unmarshal(item, &text); err == nil && text.Type == "text" {
				t.Content = append(t.Content, text)
				continue
			}
			var image ImageContent
			if err := json.Unmarshal(item, &image); err == nil && image.Type == "image" {
				t.Content = append(t.Content, image)
				continue
			}
			var resource ResourceContent
			if err := json.Unmarshal(item, &resource); err == nil && resource.URI != "" {
				t.Content = append(t.Content, resource)
				continue
			}
			// Fallback to raw interface
			var iface interface{}
			json.Unmarshal(item, &iface)
			t.Content = append(t.Content, iface)
		}
		return nil
	}

	// Handle object response with content field
	type contentWrapper struct {
		Content []interface{} `json:"content"`
		IsError bool           `json:"isError,omitempty"`
	}
	var wrapper contentWrapper
	if err := json.Unmarshal(data, &wrapper); err == nil {
		t.Content = wrapper.Content
		t.IsError = wrapper.IsError
		return nil
	}

	return nil
}

// GetText extracts text content from a tool call result
func (t *ToolCallResult) GetText() string {
	var result strings.Builder
	for _, content := range t.Content {
		if text, ok := content.(TextContent); ok {
			result.WriteString(text.Text)
		}
		if text, ok := content.(map[string]interface{}); ok {
			if typ, ok := text["type"].(string); ok && typ == "text" {
				if txt, ok := text["text"].(string); ok {
					result.WriteString(txt)
				}
			}
		}
	}
	return result.String()
}
