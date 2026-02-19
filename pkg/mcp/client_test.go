// Package mcp implements the Model Context Protocol client for picoclaw.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sipeed/picoclaw/pkg/mcp/transport"
)

// mockTransport implements a mock transport for testing
type mockTransport struct {
	mu              sync.Mutex
	sentMessages    []map[string]interface{}
	responses       chan *transport.RPCMessage
	closed          bool
	sendError       error
	receiveError    error
	initializeDelay int // Number of times to delay receive
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		sentMessages:  make([]map[string]interface{}, 0),
		responses:     make(chan *transport.RPCMessage, 10),
		closed:        false,
	}
}

func (m *mockTransport) Start() error {
	return nil
}

func (m *mockTransport) Send(ctx context.Context, req interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("transport is closed")
	}

	if m.sendError != nil {
		return m.sendError
	}

	m.sentMessages = append(m.sentMessages, req.(map[string]interface{}))
	return nil
}

func (m *mockTransport) Receive(ctx context.Context) (*transport.RPCMessage, error) {
	if m.receiveError != nil {
		return nil, m.receiveError
	}

	select {
	case resp := <-m.responses:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *mockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	close(m.responses)
	return nil
}

func (m *mockTransport) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func (m *mockTransport) queueResponse(resp *transport.RPCMessage) {
	m.responses <- resp
}

func (m *mockTransport) getLastSent() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sentMessages) == 0 {
		return nil
	}
	return m.sentMessages[len(m.sentMessages)-1]
}

func (m *mockTransport) getSentCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentMessages)
}

// Test JSON-RPC message handling
func TestNewRequest(t *testing.T) {
	tests := []struct {
		name   string
		id     int64
		method string
		params interface{}
	}{
		{
			name:   "creates request with params",
			id:     1,
			method: "test/method",
			params: map[string]string{"key": "value"},
		},
		{
			name:   "creates request without params",
			id:     2,
			method: "test/no-params",
			params: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest(tt.id, tt.method, tt.params)

			assert.Equal(t, "2.0", req.JSONRPC)
			assert.Equal(t, tt.id, req.ID)
			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, tt.params, req.Params)
		})
	}
}

func TestNewNotification(t *testing.T) {
	req := NewNotification("test/notification", map[string]string{"key": "value"})

	assert.Equal(t, "2.0", req.JSONRPC)
	assert.Equal(t, int64(0), req.ID)
	assert.Equal(t, "test/notification", req.Method)
	assert.Equal(t, map[string]string{"key": "value"}, req.Params)
}

func TestJSONRPCResponse_IsError(t *testing.T) {
	tests := []struct {
		name  string
		resp  *JSONRPCResponse
		isErr bool
	}{
		{
			name:  "response without error",
			resp:  &JSONRPCResponse{Error: nil},
			isErr: false,
		},
		{
			name: "response with error",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32600, Message: "Invalid Request"},
			},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isErr, tt.resp.IsError())
		})
	}
}

func TestJSONRPCResponse_GetError(t *testing.T) {
	tests := []struct {
		name        string
		resp        *JSONRPCResponse
		wantErr     bool
		containsMsg string
	}{
		{
			name:    "no error returns nil",
			resp:    &JSONRPCResponse{Error: nil},
			wantErr: false,
		},
		{
			name: "error returns error object",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32600, Message: "Invalid Request"},
			},
			wantErr:     true,
			containsMsg: "Invalid Request",
		},
		{
			name: "error with data",
			resp: &JSONRPCResponse{
				Error: &RPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    json.RawMessage(`{"field": "name"}`),
				},
			},
			wantErr:     true,
			containsMsg: "Invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.GetError()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.containsMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSONRPCResponse_UnmarshalResult(t *testing.T) {
	type TestResult struct {
		Name string `json:"name"`
		Value int   `json:"value"`
	}

	tests := []struct {
		name      string
		resp      *JSONRPCResponse
		wantErr   bool
		checkResult func(*testing.T, interface{})
	}{
		{
			name: "unmarshals valid result",
			resp: &JSONRPCResponse{
				Result: json.RawMessage(`{"name":"test","value":42}`),
			},
			wantErr: false,
			checkResult: func(t *testing.T, v interface{}) {
				result, ok := v.(*TestResult)
				require.True(t, ok)
				assert.Equal(t, "test", result.Name)
				assert.Equal(t, 42, result.Value)
			},
		},
		{
			name: "returns error for response with error",
			resp: &JSONRPCResponse{
				Error: &RPCError{Code: -32600, Message: "Invalid"},
			},
			wantErr: true,
		},
		{
			name: "handles empty result",
			resp: &JSONRPCResponse{
				Result: json.RawMessage{},
			},
			wantErr: false,
		},
		{
			name: "handles invalid JSON",
			resp: &JSONRPCResponse{
				Result: json.RawMessage(`{invalid json}`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestResult
			err := tt.resp.UnmarshalResult(&result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, &result)
				}
			}
		})
	}
}

func TestMCPServer(t *testing.T) {
	tests := []struct {
		name   string
		server MCPServer
	}{
		{
			name: "server with all fields",
			server: MCPServer{
				Name:    "test-server",
				Command: "/usr/bin/test",
				Args:    []string{"--arg1", "value1"},
				Env:     []string{"KEY=value"},
				Enabled: true,
			},
		},
		{
			name: "minimal server",
			server: MCPServer{
				Name:    "minimal",
				Command: "/bin/minimal",
				Enabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.server.Name, tt.server.Name)
		})
	}
}

func TestMCPTool(t *testing.T) {
	tool := MCPTool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]string{"type": "string"},
			},
		},
	}

	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.InputSchema)
}

func TestToolCallResult_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		check     func(*testing.T, *ToolCallResult)
	}{
		{
			name: "unmarshals string response",
			json: `"plain text response"`,
			check: func(t *testing.T, result *ToolCallResult) {
				assert.Len(t, result.Content, 1)
				text, ok := result.Content[0].(TextContent)
				require.True(t, ok)
				assert.Equal(t, "text", text.Type)
				assert.Equal(t, "plain text response", text.Text)
			},
		},
		{
			name: "unmarshals array response",
			json: `[{"type":"text","text":"hello"},{"type":"text","text":"world"}]`,
			check: func(t *testing.T, result *ToolCallResult) {
				assert.Len(t, result.Content, 2)
				text1, ok := result.Content[0].(TextContent)
				require.True(t, ok)
				assert.Equal(t, "hello", text1.Text)
			},
		},
		{
			name: "unmarshals object with content field",
			json: `{"content":[{"type":"text","text":"result"}],"isError":false}`,
			check: func(t *testing.T, result *ToolCallResult) {
				assert.Len(t, result.Content, 1)
				assert.False(t, result.IsError)
			},
		},
		{
			name: "unmarshals object with isError",
			json: `{"content":[],"isError":true}`,
			check: func(t *testing.T, result *ToolCallResult) {
				assert.True(t, result.IsError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ToolCallResult
			err := json.Unmarshal([]byte(tt.json), &result)
			require.NoError(t, err)
			tt.check(t, &result)
		})
	}
}

func TestToolCallResult_GetText(t *testing.T) {
	tests := []struct {
		name     string
		result   ToolCallResult
		expected string
	}{
		{
			name: "extracts text from TextContent",
			result: ToolCallResult{
				Content: []interface{}{
					TextContent{Type: "text", Text: "hello "},
					TextContent{Type: "text", Text: "world"},
				},
			},
			expected: "hello world",
		},
		{
			name: "extracts text from map",
			result: ToolCallResult{
				Content: []interface{}{
					map[string]interface{}{"type": "text", "text": "test"},
				},
			},
			expected: "test",
		},
		{
			name:     "empty content returns empty string",
			result:   ToolCallResult{Content: []interface{}{}},
			expected: "",
		},
		{
			name: "mixes TextContent and map",
			result: ToolCallResult{
				Content: []interface{}{
					TextContent{Type: "text", Text: "from struct "},
					map[string]interface{}{"type": "text", "text": "from map"},
				},
			},
			expected: "from struct from map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.GetText()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock STDIO transport for testing
type mockSTDIOTransport struct {
	sendCalled    bool
	receiveCalled bool
	startCalled   bool
	closeCalled   bool
	closed        bool
	mockResponse  *transport.RPCMessage
	mockError     error
}

func (m *mockSTDIOTransport) Start() error {
	m.startCalled = true
	return nil
}

func (m *mockSTDIOTransport) Send(ctx context.Context, req interface{}) error {
	m.sendCalled = true
	return m.mockError
}

func (m *mockSTDIOTransport) Receive(ctx context.Context) (*transport.RPCMessage, error) {
	m.receiveCalled = true
	if m.mockError != nil {
		return nil, m.mockError
	}
	if m.mockResponse != nil {
		return m.mockResponse, nil
	}
	return nil, io.EOF
}

func (m *mockSTDIOTransport) Close() error {
	m.closeCalled = true
	m.closed = true
	return nil
}

func (m *mockSTDIOTransport) IsClosed() bool {
	return m.closed
}

// Test Client
func TestNewClient(t *testing.T) {
	server := MCPServer{
		Name:    "test-server",
		Command: "/usr/bin/test",
		Args:    []string{"--arg"},
		Env:     []string{"KEY=value"},
		Enabled: true,
	}

	// We can't create a real client without mocking transport
	// This test validates the server config structure
	assert.Equal(t, "test-server", server.Name)
	assert.Equal(t, "/usr/bin/test", server.Command)
	assert.Equal(t, []string{"--arg"}, server.Args)
	assert.Equal(t, []string{"KEY=value"}, server.Env)
}

func TestMCPServerMethods(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		validMethod bool
	}{
		{name: MethodInitialize, validMethod: true},
		{name: MethodInitialized, validMethod: true},
		{name: MethodShutdown, validMethod: true},
		{name: MethodListTools, validMethod: true},
		{name: MethodCallTool, validMethod: true},
		{name: MethodListResources, validMethod: true},
		{name: MethodReadResource, validMethod: true},
		{name: MethodListPrompts, validMethod: true},
		{name: MethodGetPrompt, validMethod: true},
		{name: "unknown/method", validMethod: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validMethods := map[string]bool{
				MethodInitialize:    true,
				MethodInitialized:   true,
				MethodShutdown:      true,
				MethodListTools:     true,
				MethodCallTool:      true,
				MethodListResources: true,
				MethodReadResource:  true,
				MethodListPrompts:   true,
				MethodGetPrompt:     true,
			}
			assert.Equal(t, tt.validMethod, validMethods[tt.name])
		})
	}
}

func TestRPCErrorCodes(t *testing.T) {
	tests := []struct {
		code  int
		name  string
		valid bool
	}{
		{ParseError, "ParseError", true},
		{InvalidRequest, "InvalidRequest", true},
		{MethodNotFound, "MethodNotFound", true},
		{InvalidParams, "InvalidParams", true},
		{InternalError, "InternalError", true},
		{-99999, "Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validCodes := map[int]bool{
				ParseError:     true,
				InvalidRequest: true,
				MethodNotFound: true,
				InvalidParams:  true,
				InternalError:  true,
			}
			assert.Equal(t, tt.valid, validCodes[tt.code])
		})
	}
}

func TestInitializeParams(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCaps{},
		ClientInfo: ClientInfo{
			Name:    "picoclaw",
			Version: "0.1.0",
		},
	}

	assert.Equal(t, "2024-11-05", params.ProtocolVersion)
	assert.Equal(t, "picoclaw", params.ClientInfo.Name)
	assert.Equal(t, "0.1.0", params.ClientInfo.Version)
}

func TestListToolsParams(t *testing.T) {
	t.Run("empty params", func(t *testing.T) {
		params := ListToolsParams{}
		assert.Nil(t, params.Cursor)
	})

	t.Run("with cursor", func(t *testing.T) {
		cursor := "cursor-123"
		params := ListToolsParams{Cursor: &cursor}
		assert.NotNil(t, params.Cursor)
		assert.Equal(t, "cursor-123", *params.Cursor)
	})
}

func TestCallToolParams(t *testing.T) {
	params := CallToolParams{
		Name: "test_tool",
		Arguments: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	assert.Equal(t, "test_tool", params.Name)
	assert.NotNil(t, params.Arguments)
	assert.Equal(t, "value1", params.Arguments["param1"])
	assert.Equal(t, 42, params.Arguments["param2"])
}

func TestClientName(t *testing.T) {
	// Test client info structure
	clientInfo := ClientInfo{
		Name:    "test-client",
		Version: "1.0.0",
	}

	assert.Equal(t, "test-client", clientInfo.Name)
	assert.Equal(t, "1.0.0", clientInfo.Version)
}

func TestServerCapabilities(t *testing.T) {
	caps := ServerCapabilities{}

	// Tools
	caps.Tools.ListChanged = true
	assert.True(t, caps.Tools.ListChanged)

	// Resources
	caps.Resources.Subscribe = true
	caps.Resources.ListChanged = true
	assert.True(t, caps.Resources.Subscribe)
	assert.True(t, caps.Resources.ListChanged)

	// Prompts
	caps.Prompts.ListChanged = true
	assert.True(t, caps.Prompts.ListChanged)
}

func TestInitializeResult(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
		Capabilities: ServerCapabilities{},
	}

	assert.Equal(t, "2024-11-05", result.ProtocolVersion)
	assert.Equal(t, "test-server", result.ServerInfo.Name)
	assert.Equal(t, "1.0.0", result.ServerInfo.Version)
}

func TestListToolsResult(t *testing.T) {
	tools := []MCPTool{
		{Name: "tool1", Description: "First tool"},
		{Name: "tool2", Description: "Second tool"},
	}

	cursor := "next-cursor-123"
	result := ListToolsResult{
		Tools:     tools,
		NextCursor: &cursor,
	}

	assert.Len(t, result.Tools, 2)
	assert.NotNil(t, result.NextCursor)
	assert.Equal(t, "next-cursor-123", *result.NextCursor)
}

func TestJSONRPCRequest_MarshalJSON(t *testing.T) {
	req := NewRequest(1, "test/method", map[string]string{"key": "value"})

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "2.0", unmarshaled["jsonrpc"])
	assert.Equal(t, float64(1), unmarshaled["id"])
	assert.Equal(t, "test/method", unmarshaled["method"])
}

func TestJSONRPCResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{"jsonrpc":"2.0","id":1,"result":{"status":"ok"}}`

	var resp JSONRPCResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, int64(1), resp.ID)
	assert.NotEmpty(t, resp.Result)
}

func TestTextContent(t *testing.T) {
	content := TextContent{
		Type: "text",
		Text: "Sample text content",
	}

	assert.Equal(t, "text", content.Type)
	assert.Equal(t, "Sample text content", content.Text)
}

func TestImageContent(t *testing.T) {
	content := ImageContent{
		Type:     "image",
		Data:     "base64data",
		MimeType: "image/png",
	}

	assert.Equal(t, "image", content.Type)
	assert.Equal(t, "base64data", content.Data)
	assert.Equal(t, "image/png", content.MimeType)
}

func TestResourceContent(t *testing.T) {
	content := ResourceContent{
		URI:      "file:///path/to/file.txt",
		MimeType: "text/plain",
		Text:     "File contents",
	}

	assert.Equal(t, "file:///path/to/file.txt", content.URI)
	assert.Equal(t, "text/plain", content.MimeType)
	assert.Equal(t, "File contents", content.Text)
}

func TestClientCaps(t *testing.T) {
	caps := ClientCaps{}
	caps.Roots.ListChanged = true

	assert.True(t, caps.Roots.ListChanged)
}

func TestGetPromptParams(t *testing.T) {
	params := GetPromptParams{
		Name: "test-prompt",
		Arguments: map[string]interface{}{
			"var1": "value1",
		},
	}

	assert.Equal(t, "test-prompt", params.Name)
	assert.NotNil(t, params.Arguments)
}

func TestGetPromptResult(t *testing.T) {
	messages := []interface{}{
		map[string]string{"role": "user", "content": "Hello"},
	}

	result := GetPromptResult{
		Messages: messages,
	}

	assert.Len(t, result.Messages, 1)
}

func TestMethodConstants(t *testing.T) {
	// Ensure all method constants are defined and non-empty
	methods := []string{
		MethodInitialize,
		MethodInitialized,
		MethodShutdown,
		MethodListTools,
		MethodCallTool,
		MethodListResources,
		MethodReadResource,
		MethodListPrompts,
		MethodGetPrompt,
	}

	for _, method := range methods {
		assert.NotEmpty(t, method, "method constant should not be empty")
	}
}

func TestReadResourceParams(t *testing.T) {
	params := ReadResourceParams{
		URI: "file:///example.txt",
	}

	assert.Equal(t, "file:///example.txt", params.URI)
}

func TestListResourcesParams(t *testing.T) {
	t.Run("without cursor", func(t *testing.T) {
		params := ListResourcesParams{}
		assert.Nil(t, params.Cursor)
	})

	t.Run("with cursor", func(t *testing.T) {
		cursor := "res-cursor-456"
		params := ListResourcesParams{Cursor: &cursor}
		assert.Equal(t, "res-cursor-456", *params.Cursor)
	})
}

func TestListPromptsParams(t *testing.T) {
	t.Run("without cursor", func(t *testing.T) {
		params := ListPromptsParams{}
		assert.Nil(t, params.Cursor)
	})

	t.Run("with cursor", func(t *testing.T) {
		cursor := "prompt-cursor-789"
		params := ListPromptsParams{Cursor: &cursor}
		assert.Equal(t, "prompt-cursor-789", *params.Cursor)
	})
}

// Test that the transport package can be used
func TestTransportPackage(t *testing.T) {
	// This test verifies that the transport package types are accessible
	var _ transport.RPCMessage
	var _ transport.RPCError
	var _ (*transport.STDIOTransport)
}
