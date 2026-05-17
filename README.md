# Kube-MCP

**Kube-MCP** is an open-source Model Context Protocol (MCP) server that exposes a Kubernetes cluster as a structured set of AI-callable tools. It enables LLM agents (like Claude, GPT-4, and local models) to introspect, diagnose, and eventually remediate cluster issues through natural language—without requiring users to know `kubectl` commands or complex YAML schemas.

## Overview

Kubernetes operations demand deep expertise. Diagnosing a crashlooping pod requires cross-referencing pod events, container logs, resource limits, image pull status, and node pressure. Kube-MCP bridges the gap between LLM reasoning and Kubernetes by exposing your cluster's API as highly optimized, token-efficient MCP tools.

### Architecture

```text
┌─────────────────────────┐        ┌─────────────────────────┐        ┌─────────────────────────┐
│       LLM Client        │        │      Kube-MCP Server    │        │  Kubernetes API Server  │
│ (Claude, Cursor, Cline) │◀──────▶│     (stdio transport)   │◀──────▶│                         │
└─────────────────────────┘  MCP   └─────────────────────────┘  REST  └─────────────────────────┘
```

Kube-MCP translates MCP JSON-RPC tool calls into strongly-typed Kubernetes Client-Go API requests. It intentionally trims heavy K8s manifests into lightweight JSON summaries so LLMs can process them without blowing up their context window.

---

## Features (M0 & M1)

Kube-MCP v1 provides a foundation tailored for robust, token-efficient read access to your cluster. 

### Available Tools

| Category | Tool | Description |
| :--- | :--- | :--- |
| **Pods** | `list_pods` | Lists pods with status, restart counts, node placements, and age. Filterable by namespace. |
| | `get_pod` | Fetches the detailed spec and status of a specific pod. |
| | `get_logs` | Streams tailing logs from a specific container. Limits output to avoid token bloat. Includes `previous` flag for crashloop diagnostics. |
| **Nodes** | `list_nodes` | Lists all nodes with readiness status, roles, kernel, and runtime versions. |
| | `describe_node` | Fetches node capacity, allocatable resources, taints, and conditions. |
| **Workloads** | `list_deployments` | Lists deployments alongside their replica status, availability, and active images. |
| | `get_deployment` | Retrieves full deployment specs including rollout strategy. |
| | `check_rollout_status` | Provides an instant snapshot of a Deployment's rollout progress. |
| **Networking** | `list_services` | Lists services showing Type, ClusterIP, and mapped ports. |
| | `list_ingresses` | Lists ingress rules, mapped hosts, and TLS status. |
| **Storage & Config**| `get_pvc_status` | Shows PersistentVolumeClaim bindings, requested capacity, and storage classes. |
| | `list_configmaps` | Lists ConfigMap names and their keys (values are truncated by default for safety). |
| | `list_namespaces` | Lists all namespaces, statuses, and ages. |
| **Events** | `get_events` | Fetches namespace-scoped events. Can be filtered by reason (e.g., `Failed`, `Evicted`). |
| **Metrics** | `get_resource_usage`| Queries `metrics-server` for top CPU/Memory consumers. Supports cluster-wide summaries or drill-downs for specific nodes/pods. |

---

## Prerequisites

1. **Go 1.23+** installed on your system.
2. A valid **KUBECONFIG** file or an active in-cluster ServiceAccount.
3. *(Optional but Recommended)* `metrics-server` installed in your cluster for `get_resource_usage` to function.

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/Aamod007/Kube-MCP.git
cd Kube-MCP
go build -o kubemcp ./cmd/server
```

You can place the resulting `kubemcp` binary anywhere in your path (e.g., `/usr/local/bin/kubemcp`).

---

## Supported Platforms & Usage

Because Kube-MCP is built on the standard **Model Context Protocol (MCP)** using the `stdio` transport, it works **out-of-the-box** with any agent, editor, or CLI tool that supports the MCP standard.

Supported platforms include:
- **Claude Desktop** & **Claude Code**
- **Cursor**
- **VS Code + GitHub Copilot** & **Copilot CLI**
- **Cline**
- **OpenCode**, **OpenClaw**, **Codex**
- **Antigravity**, **Gemini CLI**, **Pi Agent**, **Vibe CLI**, **Hermes**, **KIMI CLI**

### Configuration Examples

Below are standard configuration patterns for popular platforms. Be sure to replace `/absolute/path/to/kubemcp` with the actual path to your built binary, and update the `KUBECONFIG` path for your environment.

#### 1. Claude Desktop
Add to your configuration file (typically `~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kubemcp": {
      "command": "/absolute/path/to/kubemcp",
      "env": {
        "KUBECONFIG": "/Users/yourname/.kube/config"
      }
    }
  }
}
```

#### 2. Cursor
In Cursor, go to **Settings > Features > MCP**, click **Add New MCP Server**:
- **Name:** `KubeMCP`
- **Type:** `command`
- **Command:** `/absolute/path/to/kubemcp`
*Note: Ensure your terminal environment variables or Cursor environment is properly exposing `KUBECONFIG`.*

#### 3. Cline
Add to your `cline_mcp_settings.json` file:

```json
{
  "mcpServers": {
    "kubemcp": {
      "command": "/absolute/path/to/kubemcp",
      "env": {
        "KUBECONFIG": "/Users/yourname/.kube/config"
      }
    }
  }
}
```

### Example AI Prompts

Once configured, simply ask your AI agent:
- *"List the pods in the default namespace."*
- *"Why is the nginx deployment failing to start? Check the logs."*
- *"Show me the top 5 pods using the most memory across the cluster."*
- *"Find all Failed events in the kube-system namespace."*

---

## Development & Contributing

Kube-MCP is built using Go. The internal folder structure is organized as follows:
- `cmd/server`: The main application entrypoint.
- `internal/k8s`: Kubernetes Client-Go and `metrics-server` initialization wrappers.
- `internal/tools`: The definitions and implementations of all MCP tools.

To run tests and formatting locally:
```bash
go fmt ./...
go vet ./...
go build ./...
```

## Roadmap

Upcoming releases (M2 - M6) will introduce:
- Write tools (read/write mode toggles) for `kubectl apply`, resource deletion, and strategic patching.
- HTTP Server-Sent Events (SSE) transport implementation.
- An example standalone Go-based agentic diagnostic loop.
- Docker container publishing and Helm charts for in-cluster deployment.
