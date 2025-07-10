# kubelet-wuhrai MCP Client Integration

## Multi-Server Orchestration for Security Automation

The MCP (Model Context Protocol) Client feature enables kubelet-wuhrai to coordinate multiple specialized tools through natural language commands. This integration demonstrates automated security workflows that combine RBAC scanning with email reporting.

**Problem**: Traditional security audits require manual execution of multiple tools, data correlation, and report distribution—a time-consuming process prone to human error.

**Solution**: Single command orchestration across multiple MCP servers:

```bash
kubelet-wuhrai --mcp-client --quiet "scan rbac and send urgent report to incident-team@company.com from sender@company.com"
```

**Architecture Components:**
- **kubelet-wuhrai**: Central orchestrator interpreting natural language commands
- **Permiflow**: RBAC security scanning and analysis
- **Resend**: Automated email delivery service
- **Additional servers**: Documentation, reasoning, and extensible integrations

## Workflow Sequence Diagram

```mermaid
sequenceDiagram
    participant User
    participant kubelet-wuhrai as kubelet-wuhrai<br/>(MCP Client)
    participant Permiflow as Permiflow<br/>(MCP Server)
    participant K8s as Kubernetes<br/>Cluster
    participant Resend as Resend<br/>(MCP Server)
    participant Email as Email<br/>Recipient

    User->>kubelet-wuhrai: "scan rbac and send report to admin@company.com"
    kubelet-wuhrai->>Permiflow: scan_rbac()
    Permiflow->>K8s: Query RBAC policies
    K8s-->>Permiflow: Return roles, bindings, permissions
    Permiflow->>Permiflow: Analyze security risks
    Permiflow-->>kubelet-wuhrai: Security findings report
    kubelet-wuhrai->>kubelet-wuhrai: Format report for email
    kubelet-wuhrai->>Resend: send_email(to, from, subject, content)
    Resend->>Email: Deliver formatted security report
    Email-->>User: Email confirmation
    kubelet-wuhrai-->>User: "✅ RBAC scan completed and report sent"
```

## Execution Flow

The command execution follows this sequence:
1. **kubelet-wuhrai** parses the natural language request
2. [**Permiflow**](https://github.com/tutran-se/permiflow) performs comprehensive RBAC analysis across cluster resources
4. [**Resend**](https://github.com/resend/mcp-send-email) formats and delivers the security report via email

**Extensibility**: The architecture supports additional MCP servers for Slack notifications, Jira ticket creation, compliance databases, and custom integrations.

## Configuration and Setup

### MCP Server Configuration

Configure the MCP servers in `~/.config/kubelet-wuhrai/mcp.yaml`:

```yaml
Servers:
- Args:
  - '~/mcp-send-email/build/index.js'
  env:
    RESEND_API_KEY: "api-key-here"
  Command: node
  Name: resend
- Name: permiflow
  URL: http://localhost:8080/mcp

```

### Quick Start

```bash
# 1. Start the Permiflow MCP server
permiflow mcp --transport http --http-port 8080

# 2. Execute kubelet-wuhrai with MCP client enabled
kubelet-wuhrai --mcp-client --quiet "scan rbac and send report to admin@company.com from sec@company.com"
```

## Automation Use Cases

### Scheduled Security Monitoring

Implement automated daily security scans using cron:

```bash
# Daily RBAC audit at 9 AM
0 9 * * * kubelet-wuhrai --mcp-client --quiet "scan rbac and send daily report to admin@company.com from sec@company.com"
```

### Incident Response

Execute immediate security assessments during incidents:

```bash
kubelet-wuhrai --mcp-client --quiet "scan rbac for production namespace and send urgent report to incident-team@company.com from sec@company.com"
```

## Usage Examples

### Interactive Mode

Launch kubelet-wuhrai in interactive mode for exploratory analysis:

```bash
kubelet-wuhrai --mcp-client
>>> "scan rbac and send report to admin@company.com"
>>> "analyze RBAC for kubeflow namespace"
>>> "show me the most dangerous permissions in production"
>>> "which service accounts can access secrets across namespaces?"
```

### Direct Commands

Execute specific security queries directly:

```bash
kubelet-wuhrai --mcp-client "show wildcard permissions and suggest fixes"
```

## Extended Integration

### Additional MCP Servers

Expand the automation capabilities by adding specialized servers:

```yaml
Servers:
  - Name: slack-notifier
    URL: "https://slack-mcp.company.com/mcp"
  - Name: jira-tickets
    URL: "https://jira-mcp.company.com/mcp"
  - Name: trivy-scanner
    Command: npx
    Args: ["-y", "@aquasecurity/trivy-mcp"]
```

### Advanced Workflows

**Multi-Channel Incident Response:**
```bash
"scan rbac, create jira ticket, email security team, post to slack"
```

**Compliance Automation:**
```bash
"scan vulnerabilities, update compliance database, email leadership"
```

## Benefits

- **Unified Interface**: Single natural language interface for multiple tools
- **Automation**: Reduces manual security audit processes
- **Consistency**: Standardized security scanning and reporting
- **Extensibility**: Modular architecture supports additional integrations
- **Efficiency**: Rapid security assessment and stakeholder notification
