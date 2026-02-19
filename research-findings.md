# PicoClaw Architecture Research Findings

**Date**: 2026-02-20
**Project**: PicoClaw v0.1.2-97-g8bf61ff
**Research Focus**: Project architecture, design patterns, and implementation details

---

## Executive Summary

PicoClaw is an ultra-lightweight personal AI assistant written in Go, designed to run on hardware with as little as 10MB RAM. The project demonstrates exceptional software engineering through its modular architecture, comprehensive tool system, and extensible integration patterns. This analysis covers the core architectural components including the agent loop, tool registry, MCP integration, and observability features.

**Key Statistics:**
- Language: Go 1.21+
- Memory Footprint: <10MB (99% smaller than OpenClaw)
- Startup Time: <1 second on 0.6GHz single-core
- Hardware Support: x86_64, ARM64, RISC-V

---

## 1. Agent Loop Implementation

### 1.1 Core Architecture

The agent loop is implemented in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/agent/loop.go` and follows a message-driven architecture with the following key components:

**AgentLoop Structure:**
```go
type AgentLoop struct {
    bus            *bus.MessageBus       // Message passing infrastructure
    cfg            *config.Config        // Configuration management
    registry       *AgentRegistry        // Agent instance management
    state          *state.Manager        // Persistent state handling
    running        atomic.Bool           // Concurrency-safe running flag
    summarizing    sync.Map              // Concurrent summarization tracking
    fallback       *providers.FallbackChain // Provider fallback management
    channelManager *channels.Manager     // Channel integration
    mcpManager     MCPManager            // MCP integration
}
```

### 1.2 Message Processing Flow

The agent implements a sophisticated message processing pipeline:

1. **Message Consumption**: Inbound messages consumed from message bus
2. **Routing**: Messages routed to appropriate agent using routing logic
3. **Session Management**: Context loaded from session history
4. **LLM Iteration Loop**: Tool-assisted LLM interactions
5. **Response Handling**: Results published to outbound bus

**Process Flow:**
```
Inbound Message → Route → runAgentLoop → runLLMIteration → Tool Execution → Response
```

### 1.3 LLM Iteration Loop

The `runLLMIteration` method implements the core ReAct (Reasoning + Acting) pattern:

```go
func (al *AgentLoop) runLLMIteration(ctx context.Context, agent *AgentInstance,
    messages []providers.Message, opts processOptions) (string, int, error)
```

**Key Features:**
- Configurable maximum iterations (default: 20)
- Automatic tool call detection and execution
- Context compression on token limit errors
- Fallback chain support for provider resilience
- Concurrent tool execution support

### 1.4 Advanced Features

**Context Management:**
- Automatic summarization when history exceeds 75% of context window
- Emergency compression dropping 50% oldest messages on limit
- Multi-part summarization for large conversations
- Token estimation using safe heuristic (2.5 chars/token for CJK support)

**Fallback Chain:**
- Multiple provider support with automatic failover
- Cooldown tracking to prevent cascading failures
- Configurable model candidates with priority ordering

---

## 2. Tool Registration and Execution

### 2.1 Tool Registry Architecture

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/tools/registry.go`:

```go
type ToolRegistry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}
```

**Thread-Safe Operations:**
- `Register(tool Tool)`: Add new tools
- `Get(name string) (Tool, bool)`: Retrieve tools
- `Execute(ctx, name, args)`: Execute with context
- `ExecuteWithContext(...)`: Enhanced execution with channel/chat context

### 2.2 Tool Interface

All tools implement the base `Tool` interface:

```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) *ToolResult
}
```

**Extended Interfaces:**
- `ContextualTool`: Tools needing channel/chat context (message, spawn)
- `AsyncTool`: Tools with asynchronous execution patterns

### 2.3 Built-in Tools

**Filesystem Tools:**
- `read_file`: Read file contents (workspace-restricted)
- `write_file`: Write/create files
- `edit_file`: Edit file with string replacement
- `append_file`: Append content to files
- `list_dir`: List directory contents

**System Tools:**
- `exec`: Execute shell commands (with safety guards)
- `cron`: Schedule and manage tasks

**Network Tools:**
- `web_search`: Brave, DuckDuckGo, Perplexity integration
- `web_fetch`: Fetch and convert web content

**Communication Tools:**
- `message`: Send messages via channels
- `spawn`: Create subagents with allowlist control

**Hardware Tools:**
- `i2c_tool`: I2C bus communication (Linux)
- `spi_tool`: SPI bus communication (Linux)

### 2.4 Tool Execution Flow

1. **Tool Lookup**: Retrieve from registry by name
2. **Context Injection**: Set channel/chat context for contextual tools
3. **Async Callback**: Inject callback for async tools if provided
4. **Execution**: Execute with timeout and error handling
5. **Logging**: Comprehensive execution logging with duration
6. **Result Processing**: Handle ForLLM and ForUser content separately

**Safety Features:**
- Workspace path validation
- Dangerous command pattern blocking
- Configurable deny patterns
- Execution timeout support

---

## 3. MCP (Model Context Protocol) Integration

### 3.1 MCP Client Architecture

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/mcp/client.go`:

```go
type Client struct {
    server       MCPServer         // Server configuration
    transport    *transport.STDIOTransport // Communication layer
    requestID    int64             // Request tracking
    capabilities ServerCapabilities // Server capabilities
    tools        []MCPTool         // Available tools
    initialized  bool              // Connection state
    mu           sync.Mutex        // Concurrency control
}
```

### 3.2 MCP Protocol Implementation

**Protocol Version:** 2024-11-05

**Supported Methods:**
- `initialize`: Handshake and capability exchange
- `initialized`: Notification completion
- `shutdown`: Graceful connection close
- `tools/list`: Discover available tools
- `tools/call`: Execute tools with arguments

### 3.3 Transport Layer

**STDIO Transport:**
- Process-based communication via stdin/stdout
- JSON-RPC 2.0 protocol
- Configurable command, arguments, and environment
- Automatic process lifecycle management

### 3.4 Tool Integration

MCP tools are automatically discovered and integrated:

```go
func (c *Client) ListTools(ctx context.Context) ([]MCPTool, error) {
    // Fetch tools from MCP server
    // Cache in client.tools
    // Return for registration
}
```

**Configuration Example:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "filesystem",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
        "enabled": true
      }
    ]
  }
}
```

### 3.5 Manager Pattern

The MCPManager handles lifecycle:
- Start all configured servers on agent startup
- Graceful shutdown on termination
- Error handling and logging
- Tool registration with agent registries

---

## 4. Observability Features

### 4.1 Structured Logging

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/logger/logger.go`:

**Log Levels:**
- DEBUG, INFO, WARN, ERROR, FATAL

**Log Entry Structure:**
```go
type LogEntry struct {
    Level     string
    Timestamp string
    Component string
    Message   string
    Fields    map[string]interface{}
    Caller    string
}
```

**Features:**
- Dual output: Console (human-readable) + File (JSON)
- Component-based logging for filtering
- Automatic caller information (file:line: function)
- Thread-safe operations with read-write mutex

### 4.2 Distributed Tracing

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/observability/tracing/trace.go`:

**Span Structure:**
```go
type Span struct {
    ID        string
    ParentID  string
    Name      string
    StartTime time.Time
    EndTime   time.Time
    Duration  int64
    TraceID   string
    Component string
    Fields    map[string]interface{}
}
```

**Tracing Features:**
- UUID-based trace and span ID generation
- Parent-child span relationships
- Component and field tagging
- Multiple recorder implementations (default, in-memory)

**Fluent API:**
```go
ctx = tracing.StartSpan(ctx, "operation").
    Component("agent").
    Field("key", "value").
    End()
```

### 4.3 Log Filtering

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/observability/filter/filter.go`:

**Filter Capabilities:**
- Filter by component name
- Filter by trace ID
- Filter by session key
- Time-based filtering
- Level-based filtering

### 4.4 Debug Management

Located in `/home/nick/home/projects/ai/agents/claw/picoclaw/pkg/observability/debug/debug.go`:

**Debug Features:**
- Per-component log level control
- Dynamic level adjustment at runtime
- Debug component registration
- Configuration persistence

---

## 5. Design Patterns and Architecture

### 5.1 Architectural Patterns

**1. Message Bus Pattern:**
- Decoupled communication between components
- Inbound/outbound message separation
- Multiple consumer support

**2. Registry Pattern:**
- Tool registration and discovery
- Agent instance management
- Dynamic extension points

**3. Strategy Pattern:**
- Provider abstraction (LLMProvider interface)
- Multiple transport implementations
- Pluggable tool execution

**4. Builder Pattern:**
- Context builder for message construction
- Span builder for tracing
- Configuration builders

**5. Chain of Responsibility:**
- Fallback chain for provider failover
- Message routing through bindings
- Error classification and handling

### 5.2 Concurrency Patterns

**1. Goroutine Usage:**
- Concurrent tool execution
- Async tool callbacks
- Background summarization

**2. Synchronization Primitives:**
- sync.RWMutex for registry access
- sync.Map for concurrent summarization tracking
- atomic.Bool for running state

**3. Context Propagation:**
- Request-scoped data throughout call chains
- Cancellation propagation
- Trace context distribution

### 5.3 Configuration Architecture

**Layered Configuration:**
1. Default values in code
2. JSON configuration file (`~/.picoclaw/config.json`)
3. Environment variables (automatic binding)
4. Command-line overrides

**Configuration Structure:**
- Agents: Multiple agent support with bindings
- Providers: Multiple LLM providers with fallback
- Channels: 10+ chat platform integrations
- Tools: Configurable tool settings
- Observability: Logging and tracing configuration
- MCP: External tool server configuration

---

## 6. Security and Sandboxing

### 6.1 Workspace Restrictions

**Default Behavior:**
- All file operations restricted to workspace
- Command execution restricted to workspace paths
- Configurable via `restrict_to_workspace`

**Protected Tools:**
- read_file, write_file, edit_file, append_file, list_dir, exec

### 6.2 Command Safety

**Blocked Patterns:**
- `rm -rf`, `del /f`, `rmdir /s` (bulk deletion)
- `format`, `mkfs`, `diskpart` (disk formatting)
- `dd if=` (disk imaging)
- `/dev/sd[a-z]` writes (direct disk access)
- `shutdown`, `reboot`, `poweroff` (system control)
- Fork bombs

### 6.3 Allowlist System

**Subagent Control:**
- Agent-scoped allowlist for spawn operations
- Per-agent subagent configuration
- Binding-based routing restrictions

---

## 7. Provider Architecture

### 7.1 Provider Types

**HTTP-Compatible (OpenAI Protocol):**
- OpenRouter, OpenAI, Groq, Zhipu, VLLM, Gemini, Nvidia, Moonshot, DeepSeek

**Anthropic Protocol:**
- Native Claude API support

**Special Providers:**
- Claude CLI: Integration with Claude CLI
- Codex CLI: OpenAI CLI integration
- GitHub Copilot: Local Copilot integration

### 7.2 Provider Resolution

**Resolution Priority:**
1. Explicit provider configuration
2. Model name inference
3. Fallback to default provider

**Model Name Patterns:**
- `zhipu`/`glm`/`zai` → Zhipu provider
- `claude`/`anthropic` → Anthropic provider
- `gpt`/`openai` → OpenAI provider
- `gemini`/`google` → Gemini provider

---

## 8. Session Management

### 8.1 Session Structure

**Components:**
- Message history (conversation)
- Summary (condensed context)
- Metadata (timestamps, tokens)

### 8.2 Session Persistence

**Storage:**
- File-based in `<workspace>/sessions/`
- JSON format for each session
- Automatic save after each interaction

### 8.3 Memory Optimization

**Summarization Triggers:**
- >20 messages in history
- >75% of context window used

**Compression Strategy:**
- Keep last 4 messages for continuity
- Multi-part summarization for large conversations
- Emergency compression when limit hit

---

## 9. Channel Integration

### 9.1 Supported Channels

**Chat Platforms:**
- Telegram (recommended, easy setup)
- Discord (bot with intents)
- QQ (Chinese platform)
- DingTalk (enterprise)
- LINE (webhook-based)
- Slack (app tokens)
- OneBot (generic protocol)

**Special Channels:**
- MaixCAM (hardware AI camera)
- WhatsApp (bridge-based)

### 9.2 Channel Manager

**Features:**
- Multiple concurrent channel support
- Unified message format
- Error handling and reconnection
- User allowlist support

---

## 10. Configuration Requirements

### 10.1 API Key Setup

**To run picoclaw with the research task, you need:**

1. **Zhipu API Key** (for glm-4.7 model):
   - Get key from: https://bigmodel.cn/usercenter/proj-mgmt/apikeys
   - Add to config: `providers.zhipu.api_key`
   - Environment: `PICOCLAW_PROVIDERS_ZHIPU_API_KEY`

2. **Alternative Providers:**
   - OpenRouter: https://openrouter.ai/keys (access to all models)
   - Anthropic: https://console.anthropic.com
   - OpenAI: https://platform.openai.com
   - Gemini: https://aistudio.google.com/api-keys

### 10.2 Configuration File

**Location:** `~/.picoclaw/config.json`

**Current Status:** Configuration file created but API key not set

**To Complete Setup:**
```bash
# Option 1: Edit config file
vim ~/.picoclaw/config.json
# Add your API key to providers.zhipu.api_key

# Option 2: Set environment variable
export PICOCLAW_PROVIDERS_ZHIPU_API_KEY="your-api-key-here"

# Then run picoclaw
./build/picoclaw-linux-amd64 agent -m "Your research task here"
```

---

## 11. Key Strengths

1. **Ultra-Lightweight**: <10MB memory footprint enables deployment on $10 hardware
2. **Modular Design**: Clean separation of concerns with well-defined interfaces
3. **Extensibility**: Easy tool addition via registry pattern
4. **MCP Integration**: Modern protocol for external tool integration
5. **Comprehensive Observability**: Structured logging and distributed tracing
6. **Multi-Provider Support**: Fallback chains for reliability
7. **Security First**: Workspace sandboxing and command safety guards
8. **Cross-Platform**: x86_64, ARM64, and RISC-V support

---

## 12. Potential Improvements

1. **Configuration Validation**: Add schema validation for config files
2. **Tool Testing**: Comprehensive test coverage for all tools
3. **Documentation**: More inline code documentation
4. **Error Recovery**: Enhanced error classification and recovery
5. **Metrics**: Prometheus/OpenTelemetry metrics export
6. **Tool Versioning**: Version compatibility checking for MCP tools
7. **Hot Reload**: Configuration changes without restart
8. **Batch Operations**: Bulk tool execution support

---

## 13. Conclusion

PicoClaw represents an impressive feat of software engineering, successfully translating complex AI agent capabilities into a minimal resource footprint. The architecture demonstrates:

- **Clean separation of concerns** through well-defined interfaces
- **Robust error handling** with fallback chains and retries
- **Modern design patterns** (registry, strategy, builder, chain of responsibility)
- **Production-ready observability** with structured logging and tracing
- **Extensible integration** through MCP and tool registry

The codebase is maintainable, testable, and ready for extension. The 10MB memory target drives smart architectural decisions that result in cleaner, more focused code. This project serves as an excellent reference implementation for lightweight AI agent systems.

---

## Appendix A: File Structure

```
/home/nick/home/projects/ai/agents/claw/picoclaw/
├── pkg/
│   ├── agent/           # Agent loop, instance, registry
│   ├── tools/           # Tool implementations and registry
│   ├── providers/       # LLM provider implementations
│   ├── config/          # Configuration management
│   ├── mcp/             # Model Context Protocol client
│   ├── observability/   # Logging, tracing, filtering
│   ├── channels/        # Chat platform integrations
│   ├── session/         # Session persistence
│   ├── routing/         # Message routing logic
│   └── bus/             # Message bus implementation
├── cmd/picoclaw/        # CLI entry point
├── config/              # Configuration examples
└── docs/                # Documentation
```

## Appendix B: Quick Start Commands

```bash
# Initialize configuration
./build/picoclaw-linux-amd64 onboard

# Interactive chat
./build/picoclaw-linux-amd64 agent

# Single message
./build/picoclaw-linux-amd64 agent -m "Your message"

# Start gateway
./build/picoclaw-linux-amd64 gateway

# View logs
./build/picoclaw-linux-amd64 logs

# Check status
./build/picoclaw-linux-amd64 status
```

---

**Research Completed**: 2026-02-20
**Analyzed By**: Manual code analysis and architectural review
**Total Files Analyzed**: 20+ core files across 8 packages
