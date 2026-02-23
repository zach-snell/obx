# obx

[![CI](https://github.com/zach-snell/obx/actions/workflows/ci.yml/badge.svg)](https://github.com/zach-snell/obx/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zach-snell/obx)](https://goreportcard.com/report/github.com/zach-snell/obx)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-zach--snell.github.io-blue)](https://zach-snell.github.io/obx/)

A fast, lightweight [MCP](https://modelcontextprotocol.io/) server for [Obsidian](https://obsidian.md/) vaults. Built in Go for speed and simplicity.

**[Documentation](https://zach-snell.github.io/obx/)** | **[Quick Start](#quick-start)** | **[MCP Tool Reference](#mcp-tool-reference-16-multiplexed)**

<p align="center">
  <img src="demo.gif" alt="obx CLI demo" width="700" />
</p>

## Why This Project?

| Feature | obx | Other MCP Servers |
|---------|-----------------|-------------------|
| **No plugins required** | Works directly with vault files | Often require Obsidian REST API plugin |
| **Single binary** | One file, zero dependencies | Node.js/Python runtime needed |
| **Cross-platform** | macOS, Linux, Windows | Often have platform issues |
| **72 actions** | 16 multiplexed tools, comprehensive vault operations | Typically 10-20 tools |
| **Fast startup** | ~10ms | Seconds for interpreted languages |

## Quick Start

**1. Install with one command:**

```bash
curl -sSL https://raw.githubusercontent.com/zach-snell/obx/main/install.sh | bash
```

This auto-detects your OS/architecture and installs to `/usr/local/bin`.

> **No sudo?** Install to `~/.local/bin` instead:
> ```bash
> curl -sSL https://raw.githubusercontent.com/zach-snell/obx/main/install.sh | bash -s -- --user
> ```

<details>
<summary>Manual download</summary>

```bash
# macOS (Apple Silicon)
curl -sSL https://github.com/zach-snell/obx/releases/latest/download/obx-darwin-arm64 -o obx && chmod +x obx

# macOS (Intel)
curl -sSL https://github.com/zach-snell/obx/releases/latest/download/obx-darwin-amd64 -o obx && chmod +x obx

# Linux
curl -sSL https://github.com/zach-snell/obx/releases/latest/download/obx-linux-amd64 -o obx && chmod +x obx
```
</details>

**2. Configure your MCP client:**

<details>
<summary><b>Claude Desktop</b></summary>

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obx",
      "args": ["mcp", "/path/to/your/vault"]
    }
  }
}
```
</details>

<details>
<summary><b>Claude Code</b></summary>

The server will be auto-discovered, or add to your config:

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obx",
      "args": ["mcp", "/path/to/your/vault"]
    }
  }
}
```
</details>

<details>
<summary><b>HTTP Streamable Transport</b></summary>

Run as an HTTP server for remote access or multi-client setups:

```bash
# Start HTTP server on port 8080
obx mcp /path/to/vault --http :8080
# or via env var
OBSIDIAN_ADDR=:8080 obx mcp /path/to/vault
```

Then configure your MCP client to connect to `http://localhost:8080/mcp`.
</details>

<details>
<summary><b>Other MCP Clients</b></summary>

```bash
# Run directly (communicates via stdio, default)
obx mcp /path/to/vault
```
</details>

**3. Start using it!** Ask your AI assistant to search your vault, create notes, manage tasks, etc.

> **⚠️ Paths are relative to the vault root.** All `path` parameters use paths like `projects/todo.md`, not the full filesystem path. Using absolute paths will create nested directories inside your vault.

## Installation Options

### Pre-built Binaries (Recommended)

Download from [Releases](https://github.com/zach-snell/obx/releases/latest):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `obx-darwin-arm64` |
| macOS (Intel) | `obx-darwin-amd64` |
| Linux (x64) | `obx-linux-amd64` |
| Linux (ARM) | `obx-linux-arm64` |
| Windows | `obx-windows-amd64.exe` |

### Go Install

```bash
go install github.com/zach-snell/obx/cmd/obx@latest
mv $(go env GOPATH)/bin/server $(go env GOPATH)/bin/obx
```

### Build from Source

```bash
git clone https://github.com/zach-snell/obx.git
cd obx
go build -o obx ./cmd/obx
```

### Upgrade

Just run the install script again - it always fetches the latest version:

```bash
curl -sSL https://raw.githubusercontent.com/zach-snell/obx/main/install.sh | bash
```

## Advanced Server Configuration

`obx mcp` supports flags for strict access control and dynamic operations:

### Selective Tool Disablement

If you don't want the AI assistant to access specific tools (e.g. bulk operations or deletion), you can blacklist entire tool groups using the `--disabled-tools` flag:

```bash
obx mcp /my/vault --disabled-tools manage-folders,bulk-operations,manage-frontmatter
```

### Dynamic Vault Switching

By default, an `obx mcp` instance is locked to a single vault path. If you want to allow an LLM to switch the active vault dynamically via the MCP protocol without restarting the server, enable it like this:

```bash
obx mcp /my/vault --allow-vault-switching
```

To restrict which vaults the agent is allowed to switch to, first define aliases using setup commands like `obx vault add my-notes /path/to/notes`, then pass the allowed aliases to the server:

```bash
obx mcp /my/vault --allow-vault-switching --allowed-vaults my-notes,work,personal
```

---

## MCP Tool Reference (16 Multiplexed)

`obx` multiplexes its 72 actions into 16 MCP tool groups to prevent context-window exhaustion and stay well under LLM tool limit restraints (e.g. Cursor allows 40, Copilot allows 128). You pass an `"action"` argument to each tool to route to the specific functionality.

| MCP Tool Group | Description |
|----------------|-------------|
| `manage-notes` | List, read, write, rename, append, delete, or duplicate notes. |
| `edit-note` | Perform surgical find-and-replace or precise markdown header editing. |
| `read-batch` | Read entire blocks of multiple files or extract headers simultaneously. |
| `search-vault` | Leverage fuzzy text search, regex, tags, headings, frontmatter queries, or date queries. |
| `bulk-operations` | Move directories, change root tags, or mass-update frontmatter fields across many files. |
| `manage-folders` | List, create, or recursively delete directories. |
| `manage-frontmatter` | Set, get, or remove YAML frontmatter keys; read and write Dataview inline fields. |
| `manage-links` | Resolve backlinks, forward-links, or ask the AI to suggest new graph connections. |
| `manage-tasks` | Parse lists of `- [ ]` markdown checkboxes, toggle states, or filter by completion. |
| `analyze-vault` | Hunt for broken links, orphan notes, stubs, and get massive mathematical token/word stats. |
| `manage-periodic-notes` | Fetch or instantiate Daily, Weekly, Monthly, or Yearly notes automatically. |
| `manage-templates` | Find and dynamically inject markdown blocks from your templates directory. |
| `manage-mocs` | Auto-generate alphabetical directory indices or group unlinked notes into Maps of Content. |
| `manage-canvas` | Create logic nodes and draw line edges across Obsidian JSON `.canvas` files. |
| `refactor-notes` | Split notes by heading, merge multiple notes, or extract sections to new notes. |
| `manage-vaults` | (Opt-in only) Dynamically remount the active server workspace without restarting. |

> [!NOTE]
> For the exhaustive list of `action` arguments accepted by each tool group, please read the [Official Documentation Site](https://zach-snell.github.io/obx/).

---

## Token-Efficient + Safe Writes

High-frequency tools now support compact responses and destructive tools support preview-first workflows.

### Response Modes

- `mode=compact` (default): small JSON envelope with summary + bounded data
- `mode=detailed`: legacy markdown-rich output for human reading

Example compact envelope:

```json
{
  "status": "ok",
  "mode": "compact",
  "summary": "Found 42 notes",
  "truncated": false,
  "data": {
    "total_count": 42,
    "returned_count": 42
  }
}
```

### Dry Run For Destructive/Bulk Tools

Use `dry_run=true` to preview operations without writing:

- `delete-note`, `delete-folder`
- `bulk-tag`, `bulk-move`, `bulk-set-frontmatter`
- `merge-notes`, `extract-note`, `extract-section`
- `batch-edit-note`

### Optimistic Concurrency

Write/edit tools accept optional `expected_mtime` (RFC3339Nano).  
If file modification time differs, the operation fails instead of overwriting newer changes.

---

## Usage Examples

### Daily Workflow

```
"Create today's daily note and show me my open tasks"
"What did I work on last week?"
"Find notes I haven't touched in 3 months"
```

### Research & Writing

```
"Search my vault for anything about 'machine learning'"
"Find all notes tagged #project and #active"
"What notes mention 'API design' but aren't linked?"
```

### Vault Maintenance

```
"Find orphan notes with no connections"
"Show me stub notes under 100 words"
"Generate a MOC for my projects folder"
```

### Bulk Operations

```
"Add #archive tag to all notes in the old-projects folder"
"Move all notes tagged #2023 to the archive folder"
"Set status: complete on these 5 project notes"
```

---

## Template Variables

Create templates in your `templates/` folder:

```markdown
---
title: {{title}}
date: {{date}}
status: {{status:draft}}
---

# {{title}}

Created: {{datetime}}
```

### Built-in Variables

| Variable | Example |
|----------|---------|
| `{{date}}` | `2024-01-15` |
| `{{time}}` | `14:30` |
| `{{datetime}}` | `2024-01-15 14:30` |
| `{{year}}` | `2024` |
| `{{month}}` | `01` |
| `{{day}}` | `15` |
| `{{title}}` | Note title |
| `{{filename}}` | `Note.md` |
| `{{timestamp}}` | Unix timestamp |

Use `{{var:default}}` for default values.

---

## Task Format

Compatible with [Obsidian Tasks](https://publish.obsidian.md/tasks/) plugin:

```markdown
- [ ] Open task
- [x] Completed task
- [ ] Has due date 📅 2024-01-15
- [ ] High priority ⏫
- [ ] Medium priority 🔼
- [ ] Low priority 🔽
- [ ] Tagged #project #urgent
```

---

## Security

- **Path traversal protection**: All file operations are sandboxed to your vault
- **Read-only by default**: Write operations require explicit tool calls
- **No network access**: The server only accesses local files

---

## Development

```bash
# Setup (requires Go 1.21+)
git clone https://github.com/zach-snell/obx.git
cd obx

# With mise (recommended)
mise install && mise run check

# Without mise
go build -o obx ./cmd/obx
go test -race -cover ./...
go test -bench 'Benchmark(ListNotes|SearchVault)' ./internal/vault
```

### Available Commands

| Command | Description |
|---------|-------------|
| `mise run build` | Build binary |
| `mise run test` | Run tests |
| `mise run lint` | Run linters |
| `mise run check` | All checks |
| `mise run fuzz` | Fuzz tests |

---

## FAQ

**Q: Do I need Obsidian running?**  
A: No. This server works directly with vault files on disk.

**Q: Will this conflict with Obsidian?**  
A: No. Both can access the same files safely.

**Q: What about sync (iCloud, Dropbox, etc)?**  
A: Works fine. The server reads/writes standard markdown files.

**Q: Can I use multiple vaults?**  
A: Yes! You have two main options:
1. Run multiple server instances, each pointing to a different vault on a different port.
2. Register vaults globally via `obx vault add <alias> <path>` and run the server with `obx mcp --allow-vault-switching --allowed-vaults <aliases...>`. This exposes a `manage-vaults` MCP tool allowing the AI assistant to switch between them dynamically.

---

## License

Apache 2.0 - see [LICENSE](LICENSE)
