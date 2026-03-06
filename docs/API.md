# Agent API Reference — Operator Repo

> Simplified API for Operator-OS agents. All endpoints require an API token via `Authorization: token <api-token>` header.

**Base URL:** `http://localhost:3000/api/v1`

---

## Authentication

All requests must include an API token:

```
Authorization: token agent-abc123-token
```

Tokens are scoped to an agent namespace. An agent can only access its own repositories.

---

## Repositories

### Create Repository

```
POST /user/repos
```

```json
{
  "name": "my-project",
  "description": "Agent workspace",
  "private": true,
  "auto_init": true,
  "default_branch": "main"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "my-project",
  "full_name": "agent-abc123/my-project",
  "clone_url": "http://localhost:3000/agent-abc123/my-project.git",
  "default_branch": "main"
}
```

### List Repositories

```
GET /user/repos
```

### Delete Repository

```
DELETE /repos/{owner}/{repo}
```

---

## File Operations

### Read File

```
GET /repos/{owner}/{repo}/contents/{path}?ref=main
```

**Response:**
```json
{
  "name": "hello.go",
  "path": "hello.go",
  "sha": "abc123...",
  "size": 42,
  "content": "cGFja2FnZSBtYWlu...",
  "encoding": "base64"
}
```

### Create/Update File

```
POST /repos/{owner}/{repo}/contents/{path}
```

```json
{
  "message": "add hello.go",
  "content": "cGFja2FnZSBtYWlu...",
  "branch": "main",
  "sha": "abc123..."
}
```

> `content` is base64-encoded. Include `sha` when updating an existing file.

### Delete File

```
DELETE /repos/{owner}/{repo}/contents/{path}
```

```json
{
  "message": "remove old file",
  "sha": "abc123...",
  "branch": "main"
}
```

### List Directory

```
GET /repos/{owner}/{repo}/contents/{path}?ref=main
```

Returns an array of file entries when `{path}` is a directory.

---

## Branches

### List Branches

```
GET /repos/{owner}/{repo}/branches
```

### Create Branch

```
POST /repos/{owner}/{repo}/branches
```

```json
{
  "new_branch_name": "feature-x",
  "old_branch_name": "main"
}
```

### Delete Branch

```
DELETE /repos/{owner}/{repo}/branches/{branch}
```

---

## Commits

### List Commits

```
GET /repos/{owner}/{repo}/commits?sha=main&limit=10
```

### Get Commit

```
GET /repos/{owner}/{repo}/git/commits/{sha}
```

---

## Diffs

### Compare Branches

```
GET /repos/{owner}/{repo}/compare/{base}...{head}
```

**Response includes:**
- File-level diffs with additions/deletions
- Commit list between the two refs
- Patch content per file

---

## Pull Requests

### Create Pull Request

```
POST /repos/{owner}/{repo}/pulls
```

```json
{
  "title": "Implement feature X",
  "body": "This PR adds...",
  "head": "feature-x",
  "base": "main"
}
```

### List Pull Requests

```
GET /repos/{owner}/{repo}/pulls?state=open
```

### Get Pull Request

```
GET /repos/{owner}/{repo}/pulls/{id}
```

### Merge Pull Request

```
POST /repos/{owner}/{repo}/pulls/{id}/merge
```

```json
{
  "Do": "merge",
  "merge_message_field": "Merge feature X"
}
```

`Do` options: `"merge"`, `"rebase"`, `"squash"`

### Close Pull Request

```
PATCH /repos/{owner}/{repo}/pulls/{id}
```

```json
{
  "state": "closed"
}
```

---

## Webhooks (Operator Events)

### Register Webhook

```
POST /repos/{owner}/{repo}/hooks
```

```json
{
  "type": "gitea",
  "config": {
    "url": "http://weaver:18791/hooks/git",
    "content_type": "json",
    "secret": "webhook-secret"
  },
  "events": ["push", "pull_request"],
  "active": true
}
```

### Webhook Payload Headers

| Header | Description |
|---|---|
| `X-Gitea-Event` | Event type (e.g., `push`) |
| `X-Gitea-Delivery` | Unique delivery ID |
| `X-Gitea-Signature` | HMAC-SHA256 signature |

---

## Common Patterns

### Agent Workflow: Create → Edit → Commit → PR → Merge

```bash
# 1. Create repo
curl -X POST $BASE/user/repos \
  -H "Authorization: token $TOKEN" \
  -d '{"name":"workspace","auto_init":true,"private":true}'

# 2. Write a file
curl -X POST $BASE/repos/$AGENT/workspace/contents/main.go \
  -H "Authorization: token $TOKEN" \
  -d '{"message":"initial code","content":"'$(echo -n 'package main' | base64)'"}'

# 3. Create feature branch
curl -X POST $BASE/repos/$AGENT/workspace/branches \
  -H "Authorization: token $TOKEN" \
  -d '{"new_branch_name":"feature","old_branch_name":"main"}'

# 4. Edit on branch
curl -X PUT $BASE/repos/$AGENT/workspace/contents/main.go \
  -H "Authorization: token $TOKEN" \
  -d '{"message":"update code","content":"'$(echo -n 'package main\nfunc main(){}' | base64)'","branch":"feature","sha":"<file-sha>"}'

# 5. Create PR
curl -X POST $BASE/repos/$AGENT/workspace/pulls \
  -H "Authorization: token $TOKEN" \
  -d '{"title":"Add main function","head":"feature","base":"main"}'

# 6. Merge PR
curl -X POST $BASE/repos/$AGENT/workspace/pulls/1/merge \
  -H "Authorization: token $TOKEN" \
  -d '{"Do":"squash"}'
```

### Go SDK Usage

```go
import "code.gitea.io/gitea/pkg/operator"

svc := operator.NewAgentGitService("agent-abc123", pool, config, hooks)

// Create repo
svc.CreateRepo(ctx, "workspace", operator.RepoOptions{
    Private:  true,
    AutoInit: true,
})

// Write + commit
svc.WriteFile(ctx, "workspace", "main", "hello.go", []byte("package main"))
svc.Commit(ctx, "workspace", "main", "initial commit")

// Branch + PR workflow
svc.CreateBranch(ctx, "workspace", "feature", "main")
svc.WriteFile(ctx, "workspace", "feature", "hello.go", []byte("package main\nfunc main(){}"))
svc.Commit(ctx, "workspace", "feature", "add main function")
pr, _ := svc.CreatePR(ctx, "workspace", operator.PROptions{
    Title:      "Add main function",
    HeadBranch: "feature",
    BaseBranch: "main",
})
svc.MergePR(ctx, "workspace", pr.ID, operator.MergeOptions{Strategy: "squash"})
```

---

## Error Responses

All errors return JSON:

```json
{
  "message": "repository not found",
  "url": "http://localhost:3000/api/swagger"
}
```

| Status | Meaning |
|---|---|
| `400` | Bad request (invalid parameters) |
| `401` | Invalid or missing token |
| `403` | Token valid but insufficient permissions |
| `404` | Resource not found |
| `409` | Conflict (e.g., repo already exists) |
| `422` | Validation error |
| `429` | Rate limited |
| `507` | Storage quota exceeded |

---

## Rate Limits

Default: 300 requests/minute per agent token. Configurable via `GITEA__api__MAX_RESPONSE_ITEMS`.

## Full API Documentation

The complete Gitea-compatible API is available at `/api/swagger` when the server is running. This document covers the subset most relevant to agent workflows.
