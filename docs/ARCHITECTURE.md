# Architecture — Operator Repo in Operator-OS

## Overview

Operator Repo (`go-repo`) is the version control component of Operator-OS. It gives AI agents first-class Git capabilities — creating repositories, committing code, managing branches, and merging pull requests — all through a programmatic interface optimized for machine consumption.

```
┌─────────────────────────────────────────────────┐
│                  Operator-OS                     │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │  Weaver   │  │ Identify │  │   Storage    │  │
│  │ (Orchestr)│  │  (Auth)  │  │  (Artifacts) │  │
│  └─────┬─────┘  └─────┬────┘  └──────────────┘  │
│        │              │                          │
│        ▼              ▼                          │
│  ┌─────────────────────────────┐                 │
│  │       Operator Repo         │                 │
│  │       (go-repo)             │                 │
│  │                             │                 │
│  │  ┌─────────┐ ┌──────────┐  │                 │
│  │  │Agent API│ │ Repo Pool│  │                 │
│  │  └────┬────┘ └─────┬────┘  │                 │
│  │       │            │       │                 │
│  │  ┌────▼────────────▼────┐  │                 │
│  │  │   Gitea Core Engine  │  │                 │
│  │  │   (Git + SQLite)     │  │                 │
│  │  └──────────────────────┘  │                 │
│  └─────────────────────────────┘                 │
└─────────────────────────────────────────────────┘
```

## Design Principles

### 1. Agent-First, Not Human-First

Traditional Git hosting platforms optimize for human developers: rich UIs, social features, notification systems. Operator Repo inverts this — every design decision prioritizes machine consumption:

- **API tokens over OAuth flows** — No browser-based authentication
- **Programmatic file operations** — WriteFile/ReadFile over git CLI
- **Structured responses** — JSON over rendered HTML
- **Event hooks** — Webhooks over email notifications

### 2. Single Binary, Zero Dependencies

Operator Repo ships as a single Go binary with embedded SQLite. No PostgreSQL, no Redis, no external services required. This aligns with Operator-OS's philosophy of minimal infrastructure:

- **Database:** SQLite (embedded, zero-config)
- **Storage:** Local filesystem with volume mounts
- **Auth:** API tokens validated locally
- **Search:** Built-in (no Elasticsearch)

### 3. Multi-Tenant by Default

Every agent operates in an isolated namespace managed by the `RepoPool`:

```
/data/git/
├── agent-abc123/          # Agent namespace
│   ├── project-alpha/     # Repository
│   └── project-beta/
├── agent-def456/
│   └── experiment-1/
└── _system/               # System repositories
```

Quotas, storage limits, and retention policies are enforced per-namespace based on the agent's tier (free/standard/pro/unlimited).

### 4. Event-Driven Integration

All repository operations emit events via the `HookManager`:

```
Agent commits code → EventCommitPushed → Webhook POST → Weaver orchestrator
Agent creates PR   → EventPRCreated    → Webhook POST → Review agent
Quota at 80%       → EventQuotaWarning → Webhook POST → Admin dashboard
```

This enables the broader Operator-OS ecosystem to react to code changes without polling.

## Component Map

### `pkg/operator/agent.go` — Agent Git Service

The primary interface agents interact with. Provides high-level operations that map to common Git workflows:

| Method | Description |
|---|---|
| `CreateRepo` | Initialize a new repository |
| `WriteFile` + `Commit` | Stage and commit changes |
| `CreateBranch` | Branch for parallel work |
| `CreatePR` + `MergePR` | Code review workflow |
| `Diff` | Compare branches or commits |
| `ListFiles` / `ReadFile` | Explore repository contents |

### `pkg/operator/pool.go` — Repository Pool

Manages the multi-tenant layer:
- Namespace provisioning/destruction
- Quota checking and enforcement
- Storage usage tracking
- Periodic cleanup of inactive repos

### `pkg/operator/config.go` — Configuration

Environment-variable-driven configuration with sensible defaults:
- Tier definitions (free/standard/pro/unlimited)
- Storage limits and retention policies
- Feature toggles (WebUI, SSH, API)

### `pkg/operator/hooks.go` — Event System

Webhook delivery with:
- HMAC-SHA256 payload signing
- Configurable event filters per target
- Exponential backoff retries
- Delivery history tracking

## Integration Points

### With Weaver (Orchestrator)

Weaver spawns agents that need Git access. The flow:

1. Weaver provisions an agent namespace via `RepoPool.ProvisionNamespace()`
2. Agent receives an API token scoped to its namespace
3. Agent performs Git operations via `AgentGitService`
4. Commit/PR events flow back to Weaver via webhooks
5. On agent termination, Weaver triggers cleanup or archival

### With Identify (Auth/SSO)

Operator Repo delegates external authentication to the Identify service:

- Agent tokens are issued by Identify
- Token validation happens locally (JWT verification)
- No OAuth/social login — purely token-based

### With Storage (Artifacts)

Large binary artifacts go to Operator Storage, not Git:

- Go-repo stores source code and configuration
- Operator Storage handles build artifacts, models, datasets
- Git LFS is available but discouraged for agent workflows

## Deployment

### Recommended: Docker

```bash
docker run -d \
  -p 3000:3000 \
  -v operator-repo-data:/data \
  -e OPERATOR_DEFAULT_TIER=standard \
  -e OPERATOR_WEBHOOK_SECRET=your-secret \
  --name operator-repo \
  operatoronline/go-repo
```

### Resource Requirements

| Tier | CPU | Memory | Disk |
|---|---|---|---|
| Development | 0.5 cores | 256 MB | 1 GB |
| Production (small) | 1 core | 512 MB | 10 GB |
| Production (large) | 2 cores | 1 GB | 100 GB |

### Data Layout

```
/data/
├── git/           # Repository storage (bare repos)
├── db/
│   ├── gitea.db   # Gitea's SQLite database
│   └── operator-repo.db  # Operator-specific state
├── lfs/           # Large file storage (optional)
└── tmp/           # Temporary operations
```

## Future Directions

- **Stripping:** Remove unused Gitea modules (see [STRIPPING.md](../STRIPPING.md)) to reduce binary from ~100MB to ~30MB
- **Native protocol:** Direct gRPC interface bypassing HTTP for intra-OS communication
- **Replication:** Multi-node repo mirroring for high-availability deployments
- **Smart caching:** In-memory file tree cache for frequent read operations
