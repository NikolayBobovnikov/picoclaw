// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
)

// JSONRPCRequest is a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error codes as per JSON-RPC 2.0 and MCP specification
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP Method names
const (
	MethodInitialize     = "initialize"
	MethodInitialized    = "notifications/initialized"
	MethodShutdown       = "shutdown"
	MethodListTools      = "tools/list"
	MethodCallTool       = "tools/call"
	MethodListResources  = "resources/list"
	MethodReadResource   = "resources/read"
	MethodListPrompts    = "prompts/list"
	MethodGetPrompt      = "prompts/get"
)

// NewRequest creates a new JSON-RPC request
func NewRequest(id int64, method string, params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewNotification creates a new JSON-RPC notification (no response expected)
func NewNotification(method string, params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      0, // Notifications use id:0 or omit id
		Method:  method,
		Params:  params,
	}
}

// MarshalJSON implements json.Marshaler
func (r *JSONRPCRequest) MarshalJSON() ([]byte, error) {
	type alias JSONRPCRequest
	return json.Marshal((*alias)(r))
}

// UnmarshalJSON implements json.Unmarshaler
func (r *JSONRPCResponse) UnmarshalJSON(data []byte) error {
	type alias JSONRPCResponse
	tmp := (*alias)(r)
	return json.Unmarshal(data, tmp)
}

// IsError returns true if the response contains an error
func (r *JSONRPCResponse) IsError() bool {
	return r.Error != nil
}

// GetError returns the error as a Go error
func (r *JSONRPCResponse) GetError() error {
	if r.Error == nil {
		return nil
	}
	msg := r.Error.Message
	if len(r.Error.Data) > 0 {
		var data interface{}
		if err := json.Unmarshal(r.Error.Data, &data); err == nil {
			msg = fmt.Sprintf("%s: %v", msg, data)
		}
	}
	return errors.New(msg)
}

// UnmarshalResult unmarshals the result into the provided interface
func (r *JSONRPCResponse) UnmarshalResult(v interface{}) error {
	if r.IsError() {
		return r.GetError()
	}
	if len(r.Result) == 0 {
		return nil
	}
	return json.Unmarshal(r.Result, v)
}

// InitializeParams are parameters for initialize request
type InitializeParams struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    ClientCaps `json:"capabilities"`
	ClientInfo      ClientInfo `json:"clientInfo"`
}

// ClientCaps describes client capabilities
type ClientCaps struct {
	Roots struct {
		ListChanged bool `json:"listChanged"`
	} `json:"roots"`
	Sampling struct{} `json:"sampling"` // Reserved for future use
}

// ListToolsParams are parameters for tools/list
type ListToolsParams struct {
	Cursor *string `json:"cursor,omitempty"`
}

// ListToolsResult is the result of tools/list
type ListToolsResult struct {
	Tools []MCPTool `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// CallToolParams are parameters for tools/call
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ListResourcesParams are parameters for resources/list
type ListResourcesParams struct {
	Cursor *string `json:"cursor,omitempty"`
}

// ListResourcesResult is the result of resources/list
type ListResourcesResult struct {
	Resources []interface{} `json:"resources"` // Resource objects
	NextCursor *string `json:"nextCursor,omitempty"`
}

// ReadResourceParams are parameters for resources/read
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ListPromptsParams are parameters for prompts/list
type ListPromptsParams struct {
	Cursor *string `json:"cursor,omitempty"`
}

// ListPromptsResult is the result of prompts/list
type ListPromptsResult struct {
	Prompts []interface{} `json:"prompts"` // Prompt objects
	NextCursor *string `json:"nextCursor,omitempty"`
}

// GetPromptParams are parameters for prompts/get
type GetPromptParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// GetPromptResult is the result of prompts/get
type GetPromptResult struct {
	Messages []interface{} `json:"messages"` // Prompt messages
}
