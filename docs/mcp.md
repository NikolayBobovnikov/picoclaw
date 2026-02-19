# Model Context Protocol (MCP) Integration

PicoClaw supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io), an open standard for connecting AI assistants to external tools and data sources.

## Overview

MCP allows PicoClaw to dynamically extend its capabilities by connecting to MCP servers that provide additional tools. This enables:

- **Filesystem Access** - Read/write files on your system
- **Database Queries** - Query PostgreSQL, SQLite, and other databases
- **GitHub Integration** - Interact with repositories, issues, and PRs
- **Custom Tools** - Any tool provided by MCP servers

## How MCP Works in PicoClaw

```
User Request
    ↓
PicoClaw Agent
    ↓
MCP Manager
    ↓
┌─────────────┬─────────────┬─────────────┐
│ MCP Server 1│ MCP Server 2│ MCP Server 3│
│ (Filesystem)│  (GitHub)   │ (PostgreSQL)│
└─────────────┴─────────────┴─────────────┘
    ↓            ↓            ↓
Tools wrapped and available to agent
```

The MCP integration:

1. Connects to configured MCP servers at startup
2. Lists available tools from each server
3. Wraps MCP tools as PicoClaw tools
4. Makes them available to the AI agent
5. Handles tool execution and response formatting

## Configuration

### Basic MCP Configuration

Add the `mcp` section to your `~/.picoclaw/config.json`:

```json
{
  "mcp": {
    "servers": [
      {
        "name": "filesystem",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"],
        "enabled": true
      }
    ]
  }
}
```

### Server Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `name` | string | Yes | Unique identifier for this MCP server |
| `command` | string | Yes | Command to run (e.g., "npx", "uvx") |
| `args` | array | No | Command-line arguments |
| `env` | array | No | Environment variables to pass to the server |
| `enabled` | bool | No | Whether to start this server (default: true) |

### Environment Variables

Server `enabled` status can be controlled via environment variables:

```bash
# Format: PICOCLAW_MCP_{NAME}_ENABLED
export PICOCLAW_MCP_FILESYSTEM_ENABLED=true
export PICOCLAW_MCP_GITHUB_ENABLED=false
```

## Common MCP Servers

### Filesystem Server

Provides file read/write access to specified directories.

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-filesystem
```

**Configuration:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "filesystem",
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-filesystem",
          "/home/user/documents",
          "/home/user/projects"
        ],
        "enabled": true
      }
    ]
  }
}
```

**Available Tools:**
- `mcp_filesystem_read_file` - Read file contents
- `mcp_filesystem_write_file` - Write to a file
- `mcp_filesystem_list_directory` - List directory contents
- `mcp_filesystem_create_directory` - Create a directory

**Usage Example:**
```
User: "Read the README.md file from my documents"

Agent: Uses mcp_filesystem_read_file tool
  → Path: /home/user/documents/README.md
  → Returns file contents
```

### GitHub Server

Interact with GitHub repositories, issues, and pull requests.

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-github
```

**Configuration:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "github",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": ["GITHUB_TOKEN=ghp_your_token_here"],
        "enabled": true
      }
    ]
  }
}
```

**Getting a GitHub Token:**
1. Go to https://github.com/settings/tokens
2. Generate a new token (classic)
3. Select required scopes:
   - `repo` (for private repositories)
   - `public_repo` (for public repositories only)

**Available Tools:**
- `mcp_github_create_issue` - Create a GitHub issue
- `mcp_github_create_pull_request` - Create a pull request
- `mcp_github_push_files` - Push files to a repository
- `mcp_github_search_issues` - Search for issues
- `mcp_github_search_repositories` - Search for repositories

**Usage Example:**
```
User: "Create an issue in sipeed/picoclaw about adding a new feature"

Agent: Uses mcp_github_create_issue tool
  → Repository: sipeed/picoclaw
  → Title: "Add new feature"
  → Body: "Description..."
  → Returns issue URL
```

### PostgreSQL Server

Query PostgreSQL databases.

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-postgres
```

**Configuration:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "postgres",
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-postgres",
          "postgresql://user:password@localhost:5432/mydb"
        ],
        "enabled": true
      }
    ]
  }
}
```

**Available Tools:**
- `mcp_postgres_query` - Execute SQL queries
- `mcp_postgres_list_tables` - List all tables
- `mcp_postgres_describe_table` - Get table schema

**Usage Example:**
```
User: "Show me all users who signed up this week"

Agent: Uses mcp_postgres_query tool
  → SQL: SELECT * FROM users WHERE created_at >= ...
  → Returns query results
```

### SQLite Server

Query SQLite databases.

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-sqlite
```

**Configuration:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "sqlite",
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-sqlite",
          "/path/to/database.db"
        ],
        "enabled": true
      }
    ]
  }
}
```

### Brave Search Server

Web search using Brave Search API.

**Installation:**
```bash
npm install -g @modelcontextprotocol/server-brave-search
```

**Configuration:**
```json
{
  "mcp": {
    "servers": [
      {
        "name": "brave-search",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-brave-search"],
        "env": ["BRAVE_API_KEY=your_brave_api_key"],
        "enabled": true
      }
    ]
  }
}
```

## Multiple Servers Configuration

You can configure multiple MCP servers simultaneously:

```json
{
  "mcp": {
    "servers": [
      {
        "name": "filesystem",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/docs"],
        "enabled": true
      },
      {
        "name": "github",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": ["GITHUB_TOKEN=ghp_xxx"],
        "enabled": true
      },
      {
        "name": "postgres",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/mydb"],
        "enabled": false
      }
    ]
  }
}
```

## Tool Naming Convention

MCP tools are automatically prefixed with `mcp_{server_name}_` to avoid naming conflicts:

| MCP Server | Tool Name | PicoClaw Tool Name |
|------------|-----------|-------------------|
| filesystem | read_file | mcp_filesystem_read_file |
| github | create_issue | mcp_github_create_issue |
| postgres | query | mcp_postgres_query |

## Usage Examples

### Example 1: Reading and Modifying Files

```
User: "Read the config.json file and update the timeout value"

Agent Process:
1. Uses mcp_filesystem_read_file
   → Reads config.json
2. Parses the JSON content
3. Modifies the timeout value
4. Uses mcp_filesystem_write_file
   → Writes updated config.json
```

### Example 2: GitHub Workflow

```
User: "Create a pull request for my changes"

Agent Process:
1. Uses mcp_github_search_repositories
   → Finds the target repository
2. Uses mcp_github_push_files
   → Pushes the changes
3. Uses mcp_github_create_pull_request
   → Creates the PR with description
```

### Example 3: Database Analysis

```
User: "Analyze the sales data and create a report"

Agent Process:
1. Uses mcp_postgres_list_tables
   → Discovers available tables
2. Uses mcp_postgres_query
   → Queries sales data
3. Uses mcp_filesystem_write_file
   → Writes analysis report to a file
```

## Troubleshooting

### MCP Server Won't Start

**Problem:** MCP server fails to start or connect.

**Solutions:**

1. **Check if the MCP server package is installed:**
   ```bash
   npx -y @modelcontextprotocol/server-filesystem --help
   ```

2. **Verify the command and arguments:**
   ```bash
   # Test the exact command from your config
   npx -y @modelcontextprotocol/server-filesystem /path/to/files
   ```

3. **Enable debug logging for MCP:**
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

4. **Check MCP-specific logs:**
   ```bash
   picoclaw logs --component mcp.client
   ```

### Tools Not Appearing

**Problem:** MCP server connects but tools aren't available.

**Solutions:**

1. **Check if tools were successfully registered:**
   ```bash
   picoclaw logs --component mcp.manager | grep "Registered MCP tool"
   ```

2. **Verify the server initialized correctly:**
   ```bash
   picoclaw logs --component mcp.client | grep "initialized"
   ```

3. **Check agent startup info:**
   When starting the gateway, look for:
   ```
   ✓ Agent Status:
     • Tools: 52 loaded
   ```

### Permission Errors

**Problem:** Filesystem server can't access files.

**Solutions:**

1. **Check the allowed paths in your config:**
   ```json
   {
     "args": [
       "@modelcontextprotocol/server-filesystem",
       "/home/user/documents",  // Make sure this path is correct
       "/home/user/projects"
     ]
   }
   ```

2. **Verify file permissions:**
   ```bash
   ls -la /home/user/documents
   ```

3. **Ensure PicoClaw has read/write access:**
   ```bash
   # Test with a simple file write
   echo "test" > /home/user/documents/test.txt
   ```

### GitHub Authentication Issues

**Problem:** GitHub server returns authentication errors.

**Solutions:**

1. **Verify your token has the correct scopes:**
   - Go to https://github.com/settings/tokens
   - Ensure `repo` scope is checked

2. **Test the token manually:**
   ```bash
   curl -H "Authorization: token ghp_your_token" https://api.github.com/user
   ```

3. **Check the environment variable is set:**
   ```bash
   picoclaw logs | grep "GITHUB_TOKEN"
   ```

### PostgreSQL Connection Issues

**Problem:** Can't connect to PostgreSQL database.

**Solutions:**

1. **Test the connection string:**
   ```bash
   psql "postgresql://user:password@localhost:5432/mydb"
   ```

2. **Check if PostgreSQL is running:**
   ```bash
   sudo systemctl status postgresql
   ```

3. **Verify database credentials:**
   ```bash
   psql -U user -d mydb -h localhost
   ```

## Best Practices

### Security

1. **Use environment variables for sensitive data:**
   ```json
   {
     "env": [
       "GITHUB_TOKEN=ghp_xxx",
       "DATABASE_URL=postgresql://user:pass@localhost/db"
     ]
   }
   ```

2. **Limit filesystem access:**
   Only specify directories that need to be accessed:
   ```json
   {
     "args": [
       "@modelcontextprotocol/server-filesystem",
       "/home/user/safe-folder"  // Not: /home/user
     ]
   }
   ```

3. **Use read-only database users when possible:**
   ```sql
   CREATE USER picoclaw_readonly WITH PASSWORD 'password';
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO picoclaw_readonly;
   ```

### Performance

1. **Disable unused MCP servers:**
   ```json
   {
     "enabled": false  // For servers not currently needed
   }
   ```

2. **Use specific paths for filesystem access:**
   More specific paths = faster file operations

3. **Monitor MCP server performance:**
   ```bash
   picoclaw logs --component mcp.client --level warn
   ```

### Organization

1. **Group related files:**
   Keep project files in organized directories

2. **Use descriptive server names:**
   ```json
   {
     "name": "project-filesystem"  // Not: "fs1"
   }
   ```

3. **Document your MCP setup:**
   Keep a README explaining your MCP configuration

## Advanced Configuration

### Custom MCP Servers

You can create your own MCP servers. See the [MCP documentation](https://modelcontextprotocol.io) for details.

### Using Python MCP Servers

For Python-based MCP servers, use `uvx`:

```json
{
  "name": "python-mcp-server",
  "command": "uvx",
  "args": ["my-mcp-server-package"],
  "enabled": true
}
```

### Docker Compose Integration

Run PicoClaw with MCP servers using Docker Compose:

```yaml
version: '3.8'
services:
  picoclaw:
    image: sipeed/picoclaw:latest
    environment:
      - PICOCLAW_MCP_FILESYSTEM_ENABLED=true
    volumes:
      - ./config:/root/.picoclaw
      - /home/user/documents:/documents

  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
```

## Monitoring MCP Operations

Enable debug logging to see MCP operations:

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

Example debug output:

```
[DEBUG] mcp.manager: Connecting to MCP server: filesystem
[DEBUG] mcp.client: Initializing MCP server: filesystem
[INFO] mcp.client: Server filesystem has 4 tools
[DEBUG] mcp.manager: Registered MCP tool: mcp_filesystem_read_file
[DEBUG] mcp.manager: Registered MCP tool: mcp_filesystem_write_file
[INFO] mcp.manager: MCP manager started with 1 servers
```

## Additional Resources

- [MCP Official Documentation](https://modelcontextprotocol.io)
- [MCP Server Repositories](https://github.com/modelcontextprotocol)
- [PicoClaw GitHub Issues](https://github.com/sipeed/picoclaw/issues)
- [MCP Quick Start Guide](https://modelcontextprotocol.io/quickstart)
