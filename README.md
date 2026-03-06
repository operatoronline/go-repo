<p align="center">
  <b>⚡ Operator Repo</b><br/>
  <i>Lightweight, self-hosted Git for AI agents</i>
</p>

<p align="center">
  <a href="https://github.com/operatoronline/go-repo/actions"><img src="https://github.com/operatoronline/go-repo/workflows/CI/badge.svg" alt="CI"></a>
  <a href="https://github.com/operatoronline/go-repo/releases"><img src="https://img.shields.io/github/v/release/operatoronline/go-repo" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://hub.docker.com/r/operatoronline/go-repo"><img src="https://img.shields.io/docker/image-size/operatoronline/go-repo" alt="Docker"></a>
</p>

---

**Operator Repo** (`go-repo`) is the Git service component of [Operator-OS](https://github.com/operatoronline) — an ultra-efficient AI agent operating system. It provides AI agents with lightweight, self-hosted Git version control optimized for programmatic access.

Built on [Gitea](https://gitea.io) (MIT License), Operator Repo strips away human-facing UI complexity and adds a purpose-built agent integration layer for automated repository management, multi-tenant isolation, and event-driven workflows.

## Why Operator Repo?

| Feature | Operator Repo | Standard Gitea |
|---|---|---|
| **Target user** | AI agents | Human developers |
| **Auth model** | API tokens | OAuth/social/LDAP |
| **Binary size** | ~30MB (stripped) | ~100MB |
| **Agent SDK** | Built-in Go API | REST only |
| **Multi-tenant** | Per-agent namespacing | Organization-based |
| **Storage** | SQLite embedded | Multi-DB |

## Quick Start

### Docker (Recommended)

```bash
docker build -f Dockerfile.operator -t operator-repo .
docker run -d \
  -p 3000:3000 \
  -v operator-repo-data:/data \
  -e OPERATOR_STORAGE_LIMIT=1GB \
  -e OPERATOR_MAX_REPOS=100 \
  --name operator-repo \
  operator-repo
```

### From Source

```bash
# Requires Go 1.23+
make build
./gitea --config custom/conf/app.ini
```

### Agent API Usage

```go
import "code.gitea.io/gitea/pkg/operator"

client := operator.NewAgentClient("http://localhost:3000", "agent-token-here")

// Create a repo
repo, err := client.CreateRepo(ctx, "my-project", operator.RepoOptions{
    Private: true,
})

// Commit files
err = client.WriteFile(ctx, "my-project", "main", "hello.go", []byte("package main"))
err = client.Commit(ctx, "my-project", "main", "initial commit")
```

## Architecture

Operator Repo is designed to run as a single binary with embedded SQLite — no external database required. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details on how it fits into the Operator-OS ecosystem.

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — System design and Operator-OS integration
- [Agent API](docs/API.md) — Simplified API reference for agent consumers
- [Stripping Guide](STRIPPING.md) — Analysis of removable modules for size optimization

## Project Structure

```
pkg/operator/          # Operator-OS integration layer
├── agent.go           # Agent Git interface (CreateRepo, Commit, Push, PR, etc.)
├── pool.go            # Multi-tenant repo manager
├── config.go          # Storage limits, quotas, retention policies
└── hooks.go           # Webhook/event system for agent notifications
```

## Attribution

Operator Repo is a fork of [Gitea](https://gitea.io), which is licensed under the [MIT License](LICENSE).

> Copyright (c) 2016 The Gitea Authors. All rights reserved.
> 
> Gitea is a community-managed lightweight code hosting solution written in Go.
> It is a fork of Gogs.

We gratefully acknowledge the Gitea community for building the foundation this project builds upon.

## License

[MIT](LICENSE) — Same as upstream Gitea.
