# Observability in PicoClaw

PicoClaw provides comprehensive observability features to help you monitor, debug, and understand your AI agent's behavior. This includes structured logging, distributed tracing, and flexible debug controls.

## Overview

Observability in PicoClaw consists of three main components:

1. **Structured Logging** - JSON-formatted logs with contextual fields
2. **Distributed Tracing** - Request tracking with trace IDs and span tracking
3. **Debug Mode** - Per-component log level control

## Configuration

### Basic Configuration

Add the `observability` section to your `~/.picoclaw/config.json`:

```json
{
  "observability": {
    "log_file": "/var/log/picoclaw/picoclaw.log",
    "log_level": "info",
    "enable_tracing": true,
    "trace_recorder": "default",
    "component_levels": {
      "mcp.client": "debug",
      "agent.loop": "trace"
    },
    "debug_components": ["mcp", "agent"]
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `log_file` | string | "" | Path to JSON log file (empty = stdout only) |
| `log_level` | string | "info" | Global log level: trace, debug, info, warn, error |
| `enable_tracing` | bool | true | Enable distributed tracing with trace IDs |
| `trace_recorder` | string | "default" | Span recorder type: "default" or "memory" |
| `component_levels` | map | {} | Per-component log levels |
| `debug_components` | array | [] | List of components to enable debug for |

### Environment Variables

All observability options can be set via environment variables:

```bash
export PICOCLAW_OBSERVABILITY_LOG_LEVEL=debug
export PICOCLAW_OBSERVABILITY_ENABLE_TRACING=true
export PICOCLAW_OBSERVABILITY_LOG_FILE=/var/log/picoclaw/app.log
export PICOCLAW_OBSERVABILITY_TRACE_RECORDER=memory
```

## Structured Logging

PicoClaw uses structured logging with JSON formatting for machine-readable output.

### Log Format

Each log entry includes:

```json
{
  "timestamp": "2026-02-19T15:30:45Z",
  "level": "INFO",
  "component": "agent.loop",
  "message": "Agent initialized",
  "fields": {
    "trace_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "span_id": "12345678-1234-5678-1234-567812345678",
    "session": "cli:default",
    "tools_count": 42
  },
  "caller": "agent/loop.go:125"
}
```

### Log Levels

| Level | Severity | Use Case |
|-------|----------|----------|
| `trace` | Lowest | Detailed execution flow, function entry/exit |
| `debug` | Low | Detailed debugging information |
| `info` | Medium | General informational messages |
| `warn` | High | Warning messages for potentially harmful situations |
| `error` | Critical | Error messages for critical failures |

### Component-Based Logging

Logs are tagged with component names for easy filtering:

- `agent.loop` - Main agent loop execution
- `mcp.client` - MCP client operations
- `mcp.manager` - MCP server management
- `tools.exec` - Command execution
- `tools.web` - Web search operations
- `channels.telegram` - Telegram channel
- `channels.discord` - Discord channel
- `heartbeat` - Periodic task execution
- `cron` - Scheduled job execution

## Distributed Tracing

Tracing allows you to follow a request through multiple components and services.

### Trace Context

Every request can have a unique `trace_id` that propagates through:

1. Agent loop processing
2. Tool execution
3. MCP server calls
4. Channel operations

### Enabling Tracing

```json
{
  "observability": {
    "enable_tracing": true
  }
}
```

### Trace Context in Logs

When tracing is enabled, logs automatically include:

- `trace_id` - Unique identifier for the entire request
- `span_id` - Unique identifier for the current operation
- `parent_span_id` - Parent operation's span ID

Example trace flow:

```
[INFO] agent.loop: Processing user message
  {trace_id: abc123, span_id: def456}
    ↓
[INFO] mcp.client: Calling MCP tool
  {trace_id: abc123, span_id: ghi789, parent_span_id: def456}
    ↓
[INFO] tools.web: Searching the web
  {trace_id: abc123, span_id: jkl012, parent_span_id: def456}
```

## Debug Commands

PicoClaw provides CLI commands for managing debug levels and viewing logs.

### logs Command

View and filter logs from the log file:

```bash
# View all logs
picoclaw logs

# View logs for a specific component
picoclaw logs --component mcp.client

# View logs for a specific trace
picoclaw logs --trace-id abc123-def456-...

# View logs from the last hour
picoclaw logs --since 1h

# View error logs only
picoclaw logs --level error

# Tail the log file (last 100 lines)
picoclaw logs --tail

# Get logs for a specific session
picoclaw logs --session telegram:123456
```

#### Filter Options

| Option | Description | Example |
|--------|-------------|---------|
| `--component` | Filter by component name | `--component mcp.client` |
| `--level` | Filter by log level | `--level error` |
| `--trace-id` | Filter by trace ID | `--trace-id abc123...` |
| `--session` | Filter by session key | `--session cli:default` |
| `--since` | Show logs since duration | `--since 1h` |
| `--until` | Show logs until time | `--until 2026-02-19T12:00:00Z` |
| `--tail` | Show last N lines | `--tail -n 50` |
| `--limit` | Limit number of results | `--limit 100` |

### debug Command

Manage debug levels for components:

```bash
# Set global debug level
picoclaw debug set --level debug

# Enable debug for specific components
picoclaw debug set --components mcp.client,agent.loop --level debug

# Enable trace level for a component
picoclaw debug set --components agent.loop --level trace

# Show current debug configuration
picoclaw debug show

# Reset component to default level
picoclaw debug reset --component mcp.client

# Reset all components to default
picoclaw debug reset --all
```

#### Debug Subcommands

| Subcommand | Description |
|------------|-------------|
| `set` | Set debug level for components |
| `show` | Display current debug configuration |
| `reset` | Reset component(s) to default level |

#### Debug Levels

| Level | Description |
|-------|-------------|
| `trace` | Most verbose - trace every function call |
| `debug` | Detailed debugging information |
| `info` | General information (default) |
| `warn` | Warnings only |
| `error` | Errors only |

## Usage Examples

### Debugging MCP Connections

Enable debug logging for MCP components:

```json
{
  "observability": {
    "component_levels": {
      "mcp.client": "debug",
      "mcp.manager": "debug"
    }
  }
}
```

Or via CLI:

```bash
picoclaw debug set --components mcp.client,mcp.manager --level debug
```

Then view MCP-specific logs:

```bash
picoclaw logs --component mcp.client
```

### Tracing a Request

1. Find the trace ID from any log entry:

```bash
picoclaw logs | grep "Processing user message"
```

Output:
```
[2026-02-19T15:30:45Z] [INFO] agent.loop: Processing user message
  {trace_id: abc123-def456-7890-abcd-ef1234567890, session: telegram:123456}
```

2. View all logs for that trace:

```bash
picoclaw logs --trace-id abc123-def456-7890-abcd-ef1234567890
```

### Monitoring Agent Performance

Use trace logs to measure execution time:

```json
{
  "observability": {
    "component_levels": {
      "agent.loop": "trace"
    },
    "enable_tracing": true
  }
}
```

Trace logs include duration information:

```
[TRACE] agent.loop: Completed agent iteration
  {trace_id: abc123, span_id: def456, duration_ms: 1234}
```

## Troubleshooting

### Logs Not Appearing

If logs aren't being written:

1. Check log file path permissions:
   ```bash
   ls -la /var/log/picoclaw/
   ```

2. Verify log file configuration:
   ```bash
   picoclaw status | grep observability
   ```

3. Check disk space:
   ```bash
   df -h /var/log/
   ```

### Missing Trace IDs

If trace IDs aren't appearing in logs:

1. Ensure tracing is enabled:
   ```json
   {
     "observability": {
       "enable_tracing": true
     }
   }
   ```

2. Check for conflicting trace context in your code

### Debug Mode Not Working

If debug mode doesn't seem to take effect:

1. Verify the component name is correct:
   ```bash
   picoclaw debug show
   ```

2. Check for typos in component names

3. Ensure the component actually produces logs at that level

## Best Practices

### Production Logging

For production deployments:

```json
{
  "observability": {
    "log_file": "/var/log/picoclaw/app.log",
    "log_level": "info",
    "enable_tracing": true,
    "trace_recorder": "default"
  }
}
```

### Development Debugging

For development/debugging:

```json
{
  "observability": {
    "log_level": "debug",
    "enable_tracing": true,
    "trace_recorder": "memory",
    "component_levels": {
      "agent.loop": "trace",
      "mcp.client": "debug"
    }
  }
}
```

### Log Rotation

Use logrotate or similar tools to manage log files:

```
/var/log/picoclaw/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

## Integrations

### External Logging Systems

PicoClaw's JSON logs can be easily integrated with:

- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **Datadog**
- **Splunk**
- **CloudWatch Logs**

### Example: Sending to Loki

```bash
# Install loki-log-driver
curl -s https://raw.githubusercontent.com/grafana/loki/main/cmd/promtail/promtail-docker-config.yaml -o promtail-config.yaml

# Configure promtail to read picoclaw logs
# Then run picoclaw:
picoclaw gateway 2>&1 | promtail -config.file=promtail-config.yaml
```

## API Reference

### Go API

If you're developing with PicoClaw's Go packages:

```go
import (
    "github.com/sipeed/picoclaw/pkg/observability/tracing"
    "github.com/sipeed/picoclaw/pkg/logger"
)

// Create a span
ctx, span := tracing.WithSpan(context.Background(), "operation_name")
defer tracing.End(span)

// Add fields to span
tracing.WithField(span, "user_id", "12345")

// Log with context
logger.InfoCF("component", "message", map[string]interface{}{
    "trace_id": tracing.GetTraceID(ctx),
    "span_id": tracing.GetSpanID(ctx),
})
```

See the API documentation for full details on observability interfaces.
