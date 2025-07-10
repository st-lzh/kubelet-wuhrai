# kubelet-wuhrai MCP Server

kubelet-wuhrai can run as an MCP (Model Context Protocol) server, exposing kubelet-wuhrai tools to other MCP clients. The server can run in two modes:

1. **Built-in tools only**: Exposes only kubelet-wuhrai's native tools
2. **External tool discovery**: Additionally discovers and exposes tools from other MCP servers

## Quick Start

### Basic MCP Server (Built-in tools only)

Start the MCP server with only kubelet-wuhrai's built-in tools:

```bash
kubelet-wuhrai --mcp-server
```

### Enhanced MCP Server (With external tool discovery)

Start the MCP server with external MCP tool discovery enabled:

```bash
kubelet-wuhrai --mcp-server --external-tools
```

## Configuration

The enhanced MCP server will automatically discover and expose tools from configured MCP servers when `--external-tools` is enabled. Configure MCP servers using the standard MCP client configuration.

### Example MCP Configuration

Create `~/.config/kubelet-wuhrai/mcp.yaml`:

```yaml
servers:
  filesystem:
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/files"]
  
  brave-search:
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-brave-search"]
    env:
      BRAVE_API_KEY: "your-api-key"
```

## Features

### Tool Aggregation

When external tool discovery is enabled with `--external-tools`, the kubelet-wuhrai MCP server acts as a **tool aggregator**, providing:

- All kubelet-wuhrai built-in tools (kubectl, cluster analysis, etc.)
- Tools from external MCP servers (filesystem, web search, etc.)
- Unified interface for all tools through a single MCP endpoint

### Graceful Degradation

The server handles external MCP connection failures gracefully:

- If external MCP servers are unavailable, the server continues with built-in tools only
- Individual tool failures don't affect the overall server operation
- Clear logging for troubleshooting connection issues

### Example Usage in Claude Desktop

Configure Claude Desktop to use kubelet-wuhrai as an MCP server:

**Basic usage (built-in tools only):**
```json
{
  "mcpServers": {
    "kubelet-wuhrai": {
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server"]
    }
  }
}
```

**Enhanced usage (with external tools):**
```json
{
  "mcpServers": {
    "kubelet-wuhrai": {
      "command": "kubelet-wuhrai",
      "args": ["--mcp-server", "--external-tools"]
    }
  }
}
```

## Available Tools

### Built-in Tools

kubelet-wuhrai provides these native tools:

- `kubectl_apply`: Apply Kubernetes manifests
- `kubectl_get`: Get Kubernetes resources
- `kubectl_describe`: Describe Kubernetes resources
- `kubectl_logs`: Get container logs
- `kubectl_exec`: Execute commands in containers
- And more...

### External Tools (when --external-tools is enabled)

Additional tools available depend on configured MCP servers:

- **Filesystem tools**: Read/write files, list directories
- **Web search tools**: Search the internet for information  
- **Database tools**: Query databases
- **API tools**: Interact with external APIs
- **Custom tools**: Any MCP-compatible tools

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--mcp-server` | `false` | Run in MCP server mode |
| `--external-tools` | `false` | Discover and expose external MCP tools (requires --mcp-server) |
| `--kubeconfig` | `~/.kube/config` | Path to kubeconfig file |

## Architecture

```
┌─────────────────┐    ┌───────────────────┐    ┌─────────────────┐
│   MCP Client    │───▶│ kubelet-wuhrai Server │───▶│ External Tools  │
│  (Claude, etc.) │    │                   │    │ (filesystem,    │
│                 │    │ ┌───────────────┐ │    │  web search,    │
│                 │    │ │ Built-in      │ │    │  etc.)          │
│                 │    │ │ kubectl tools │ │    │                 │
│                 │    │ └───────────────┘ │    │                 │
└─────────────────┘    └───────────────────┘    └─────────────────┘
```

The kubelet-wuhrai MCP server acts as both:
- An **MCP Server** (exposing tools to clients)
- An **MCP Client** (consuming tools from other servers, when `--external-tools` is enabled)

This creates a powerful tool aggregation pattern where kubelet-wuhrai becomes a central hub for both Kubernetes operations and general-purpose tools.

## Troubleshooting

### External Tools Not Available

If external tools aren't appearing:

1. Ensure you're using both `--mcp-server` and `--external-tools` flags
2. Check MCP configuration file exists and is valid
3. Verify external MCP servers are working independently
4. Check kubelet-wuhrai logs for connection errors
5. Try running with external tools disabled to isolate issues

### Performance Considerations

- Tool discovery adds startup time (usually 2-3 seconds) when `--external-tools` is enabled
- Each external tool call has network overhead
- Consider running without `--external-tools` for faster startup if external tools aren't needed

### Debugging

Enable verbose logging to troubleshoot:

```bash
kubelet-wuhrai --mcp-server --external-tools -v=2
```

This will show:
- MCP server connection attempts
- Tool discovery results
- Tool call routing decisions
