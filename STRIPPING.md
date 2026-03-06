# STRIPPING.md — Module Removal Analysis

> **Status:** Analysis only. No code has been deleted.  
> **Goal:** Reduce binary size and attack surface by identifying modules unnecessary for agent-only Git workflows.

## Overview

Gitea's full binary is ~100MB and includes features designed for human developer workflows: wikis, project boards, package registries, social login, email notifications, and activity dashboards. Operator Repo targets AI agents that interact exclusively via API tokens and programmatic Git operations.

This document catalogs removable modules with estimated size impact and removal complexity.

---

## Module Analysis

### 1. Wiki System

| Metric | Value |
|---|---|
| **Source paths** | `services/wiki/`, `models/wiki.go`, `routers/web/repo/wiki.go`, `routers/api/v1/repo/wiki.go` |
| **Estimated code** | ~2,500 lines |
| **Binary impact** | ~1–2 MB |
| **Removal complexity** | Low |
| **Dependencies** | Markdown rendering (shared), Git operations (shared) |

**Rationale:** Agents don't use wikis. They store documentation as repo files. The wiki system adds a parallel content storage mechanism with its own Git backend that's entirely redundant for agent workflows.

---

### 2. Project Boards

| Metric | Value |
|---|---|
| **Source paths** | `services/projects/`, `models/project/`, `routers/web/repo/projects.go`, `routers/api/v1/repo/project.go` |
| **Estimated code** | ~4,000 lines |
| **Binary impact** | ~1.5–2.5 MB |
| **Removal complexity** | Medium |
| **Dependencies** | Issue system (coupled), database models |

**Rationale:** Project boards are a visual planning tool for human teams. Agents track work through their own orchestration layer (Weaver), not Kanban boards.

---

### 3. Package Registry

| Metric | Value |
|---|---|
| **Source paths** | `services/packages/`, `models/packages/`, `routers/api/packages/`, `routers/web/repo/packages.go` |
| **Estimated code** | ~8,000 lines |
| **Binary impact** | ~4–6 MB |
| **Removal complexity** | Medium |
| **Dependencies** | Storage subsystem, container registry protocols (npm, maven, nuget, cargo, etc.) |

**Rationale:** Package registry supports 20+ package formats (npm, PyPI, Maven, NuGet, Cargo, Docker, etc.). Agents use dedicated artifact storage (Operator Storage service), not Git-hosted packages. This is the single largest removable module.

**Sub-modules (all removable):**
- `models/packages/cargo/` — Cargo/Rust
- `models/packages/composer/` — PHP Composer
- `models/packages/conda/` — Conda
- `models/packages/container/` — Docker/OCI
- `models/packages/conan/` — C/C++ Conan
- `models/packages/debian/` — Debian packages
- `models/packages/helm/` — Kubernetes Helm
- `models/packages/maven/` — Java Maven
- `models/packages/npm/` — Node.js npm
- `models/packages/nuget/` — .NET NuGet
- `models/packages/pub/` — Dart/Flutter
- `models/packages/pypi/` — Python PyPI
- `models/packages/rpm/` — RPM packages
- `models/packages/rubygems/` — Ruby gems
- `models/packages/swift/` — Swift packages
- `models/packages/vagrant/` — Vagrant boxes

---

### 4. Organization Management (Complex)

| Metric | Value |
|---|---|
| **Source paths** | `services/org/`, `models/organization/`, `routers/web/org/`, `routers/api/v1/org/` |
| **Estimated code** | ~6,000 lines |
| **Binary impact** | ~2–3 MB |
| **Removal complexity** | High |
| **Dependencies** | User system, permission model, team management |

**Rationale:** Operator Repo uses flat per-agent namespacing (managed by `pkg/operator/pool.go`), not hierarchical organizations with teams and roles. However, the organization model is deeply coupled to the permission system — removal requires careful refactoring.

**Recommendation:** Phase 2 removal. Initially disable via configuration rather than code deletion.

---

### 5. OAuth/Social Login

| Metric | Value |
|---|---|
| **Source paths** | `services/oauth2_provider/`, `services/externalaccount/`, `services/auth/source/oauth2/`, `models/auth/source/oauth2/` |
| **Estimated code** | ~3,500 lines |
| **Binary impact** | ~2–3 MB (includes OAuth2 libraries) |
| **Removal complexity** | Medium |
| **Dependencies** | `github.com/markbates/goth` library, session management |

**Rationale:** Agents authenticate with API tokens. They don't log in via GitHub, Google, GitLab, etc. The OAuth2 provider functionality (Gitea acting as an OAuth2 server) is also unnecessary — Operator-OS has its own auth service (Identify/SSO).

**Libraries removed with this module:**
- `github.com/markbates/goth` + all provider sub-packages
- Various OAuth2 JWT libraries

---

### 6. Email Notifications

| Metric | Value |
|---|---|
| **Source paths** | `services/mailer/`, `models/user/email_address.go`, email templates in `templates/mail/` |
| **Estimated code** | ~5,000 lines |
| **Binary impact** | ~2–3 MB (includes SMTP libraries, HTML email rendering) |
| **Removal complexity** | Medium |
| **Dependencies** | User notification preferences, SMTP client libraries |

**Rationale:** Agents don't have email addresses. Notifications are handled via webhooks (`pkg/operator/hooks.go`) and the Operator-OS event bus. The entire SMTP subsystem is dead weight.

**Libraries removed with this module:**
- `gopkg.in/gomail.v2` or equivalent SMTP library
- HTML email template engine overhead

---

### 7. Activity Feeds & Dashboards

| Metric | Value |
|---|---|
| **Source paths** | `services/feed/`, `models/activities/`, `routers/web/feed.go`, `routers/web/user/home.go` |
| **Estimated code** | ~4,000 lines |
| **Binary impact** | ~1.5–2 MB |
| **Removal complexity** | Low |
| **Dependencies** | Action logging, database models |

**Rationale:** Activity feeds render human-readable dashboards of recent actions. Agents query specific resources via API, not scrollable feeds. The action logging table also adds continuous write pressure to the database.

---

## Summary

| Module | Code (LOC) | Binary Reduction | Complexity | Priority |
|---|---|---|---|---|
| Package Registry | ~8,000 | 4–6 MB | Medium | **P0** |
| Organization Mgmt | ~6,000 | 2–3 MB | High | P2 |
| Email/Mailer | ~5,000 | 2–3 MB | Medium | **P0** |
| Activity Feeds | ~4,000 | 1.5–2 MB | Low | **P0** |
| Project Boards | ~4,000 | 1.5–2.5 MB | Medium | P1 |
| OAuth/Social | ~3,500 | 2–3 MB | Medium | **P0** |
| Wiki System | ~2,500 | 1–2 MB | Low | **P0** |

### Total Estimated Reduction

- **Phase 1 (P0):** ~23,000 LOC removed → **~12–18 MB binary reduction**
- **Phase 2 (P1+P2):** ~10,000 LOC removed → **~5–8 MB additional reduction**
- **Combined:** ~33,000 LOC → **~17–26 MB total reduction**
- **Target binary:** ~30–35 MB (down from ~100 MB)

### Additional Size Optimizations (Not Module Removal)

| Optimization | Impact |
|---|---|
| Strip frontend assets (templates, JS, CSS) | ~15–25 MB |
| Build with `-ldflags="-s -w"` | ~5–8 MB |
| UPX compression | ~60% of remaining |
| Disable CGo (pure Go SQLite) | Simpler builds |

With full stripping + frontend removal + build flags, a **~25–30 MB** binary is achievable.

---

## Removal Strategy

1. **Feature flags first** — Disable modules via `app.ini` configuration before removing code
2. **Build tags** — Use Go build tags to exclude modules at compile time
3. **Interface boundaries** — Replace removed services with no-op implementations
4. **Test coverage** — Ensure core Git operations pass after each removal phase
5. **Never delete upstream** — Keep removals in separate commits for easy upstream sync
