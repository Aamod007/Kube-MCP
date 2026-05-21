# Kube-MCP

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP Protocol](https://img.shields.io/badge/MCP-stdio-8B5CF6)](https://modelcontextprotocol.io/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-v1.36+-326CE5?logo=kubernetes)](https://kubernetes.io/)

**Kube-MCP** is a production-grade Model Context Protocol (MCP) server that exposes Kubernetes clusters as structured, AI-callable tools. It enables LLM agents to introspect, diagnose, and remediate cluster issues through natural language—eliminating the need for `kubectl` expertise.

---

## Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Project Structure](#project-structure)
- [Development](#development)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Architecture

```
┌─────────────────────────┐        ┌─────────────────────────┐        ┌─────────────────────────┐
│       LLM Client        │        │      Kube-MCP Server    │        │  Kubernetes API Server  │
│ (Claude, Cursor, Cline) │◀──────▶│     (stdio transport)   │◀──────▶│                         │
└─────────────────────────┘  MCP   └─────────────────────────┘  REST  └─────────────────────────┘
```

Kube-MCP translates MCP JSON-RPC tool calls into strongly-typed Kubernetes Client-Go API requests. Responses are intentionally optimized into lightweight JSON summaries, preserving LLM context windows while delivering actionable cluster data.

### Key Design Principles

- **Token-Efficient**: Trimmed responses prevent context window exhaustion
- **Read-Only by Default**: Safe introspection without accidental mutations
- **Auto-Detecting Environment**: Seamlessly works in-cluster or with local kubeconfig
- **Graceful Degradation**: Continues operating even when metrics-server is unavailable

---

## Features

### Available MCP Tools

| Category | Tool | Parameters | Description |
|:---|:---|:---|:---|
| **Pods** | `list_pods` | `namespace` | List pods with status, restarts, ready count, node, and age |
| | `get_pod` | `name`, `namespace` | Fetch detailed pod specification and status |
| | `get_logs` | `pod`, `namespace`, `container`, `tail`, `previous` | Stream container logs with 50KB cap and crashloop diagnostics |
| **Nodes** | `list_nodes` | — | List nodes with readiness, roles, version, and OS |
| | `describe_node` | `name` | Fetch node capacity, allocatable, taints, and conditions |
| **Workloads** | `list_deployments` | `namespace` | List deployments with replica counts and availability |
| | `get_deployment` | `name`, `namespace` | Fetch full deployment specification |
| | `check_rollout_status` | `deployment`, `namespace` | Monitor deployment rollout progress |
| **Networking** | `list_services` | `namespace` | List services with type, ClusterIP, and ports |
| | `list_ingresses` | `namespace` | List ingresses with hosts and load balancer addresses |
| **Storage** | `get_pvc_status` | `namespace` | List PVCs with binding status, capacity, and storage class |
| **Config** | `list_configmaps` | `namespace` | List ConfigMaps with names and keys (values omitted for safety) |
| **Namespaces** | `list_namespaces` | — | List namespaces with status phase and age |
| **Events** | `get_events` | `namespace`, `reason` | Fetch namespace events with optional reason filtering |
| **Metrics** | `get_resource_usage` | `pod_name`, `namespace`, `node_name` | Query CPU/memory usage (pod, node, or cluster-wide) |

---

## Prerequisites

| Requirement | Version | Notes |
|:---|:---|:---|
| **Go** | 1.26+ | [Download](https://go.dev/dl/) |
| **Kubernetes** | 1.24+ | Valid kubeconfig or in-cluster ServiceAccount |
| **metrics-server** | Optional | Required for `get_resource_usage` tool |

---

## Installation

### Build from Source

```bash
git clone https://github.com/Aamod007/Kube-MCP.git
cd Kube-MCP
go build -o kubemcp ./cmd/server
```

### Install Binary

```bash
sudo install kubemcp /usr/local/bin/kubemcp
```

### Verify Installation

```bash
kubemcp --help
```

---

## Configuration

Kube-MCP uses the standard MCP `stdio` transport, making it compatible with any MCP-compliant client.

### Environment Variables

| Variable | Required | Default | Description |
|:---|:---|:---|:---|
| `KUBECONFIG` | No | `~/.kube/config` | Path to Kubernetes config file |

### Client Configuration

#### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "kubemcp": {
      "command": "/usr/local/bin/kubemcp",
      "env": {
        "KUBECONFIG": "/Users/yourname/.kube/config"
      }
    }
  }
}
```

#### Cursor

1. Navigate to **Settings > Features > MCP**
2. Click **Add New MCP Server**
3. Configure:
   - **Name:** `KubeMCP`
   - **Type:** `command`
   - **Command:** `/usr/local/bin/kubemcp`

#### Cline

Edit `cline_mcp_settings.json`:

```json
{
  "mcpServers": {
    "kubemcp": {
      "command": "/usr/local/bin/kubemcp",
      "env": {
        "KUBECONFIG": "/Users/yourname/.kube/config"
      }
    }
  }
}
```

### Supported Platforms

- Claude Desktop / Claude Code
- Cursor
- VS Code + GitHub Copilot
- Cline
- OpenCode / OpenClaw / Codex
- Gemini CLI / Antigravity / Pi Agent

---

## Usage Examples

Once configured, interact with your cluster using natural language:

```
List all pods in the production namespace
```

```
Why is the nginx deployment failing? Check the logs and events.
```

```
Show me the top 5 pods by memory usage across the cluster
```

```
Find all Failed events in kube-system from the last hour
```

```
Describe the node worker-01 and check its resource pressure conditions
```

---

## Project Structure

```
Kube-MCP/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── k8s/
│   │   ├── client.go            # Kubernetes client initialization
│   │   └── metrics.go           # Metrics-server API wrapper
│   └── tools/
│       ├── registry.go          # Tool registration dispatcher
│       ├── helpers.go           # Shared response utilities
│       ├── pods.go              # Pod management tools
│       ├── nodes.go             # Node management tools
│       ├── deployments.go       # Deployment & rollout tools
│       ├── services.go          # Service & ingress tools
│       ├── events.go            # Event querying tools
│       ├── config.go            # ConfigMap tools
│       ├── storage.go           # PVC status tools
│       ├── namespaces.go        # Namespace listing tools
│       └── resource.go          # CPU/memory usage tools
├── go.mod
├── go.sum
└── README.md
```

---

## Development

### Code Quality

```bash
go fmt ./...
go vet ./...
go build ./...
```

### Run Locally

```bash
go run ./cmd/server
```

### Testing

```bash
go test ./... -v
```

---

## Roadmap

| Milestone | Features | Status |
|:---|:---|:---|
| **M0/M1** | Read-only tools, stdio transport | Completed |
| **M2** | Write tools (apply, delete, patch) with read/write mode toggles | Planned |
| **M3** | HTTP SSE transport for remote deployments | Planned |
| **M4** | Agentic diagnostic loop (autonomous troubleshooting) | Planned |
| **M5** | Docker container publishing | Planned |
| **M6** | Helm charts for in-cluster deployment | Planned |

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Guidelines

- Follow conventional commit messages
- Add tests for new functionality
- Update documentation for user-facing changes
- Ensure `go fmt` and `go vet` pass before submitting

---

## Security

- **Read-Only by Default**: All tools are introspection-only
- **No Secret Exposure**: ConfigMap values are intentionally omitted
- **Log Caps**: Log output is limited to 50KB to prevent token exhaustion
- **RBAC Respected**: Operations are scoped to the configured ServiceAccount permissions

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) by Anthropic
- [Kubernetes Client-Go](https://github.com/kubernetes/client-go)
- [mcp-go](https://github.com/mark3labs/mcp-go)
