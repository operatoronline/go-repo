// Copyright 2026 The Operator-OS Authors. MIT License.

// Package operator provides the Operator-OS integration layer for go-repo.
// It exposes a high-level, agent-friendly interface for Git operations,
// multi-tenant repository management, and event-driven workflows.
package operator

import (
	"context"
	"io"
	"time"
)

// RepoOptions configures a new repository.
type RepoOptions struct {
	// Description is a short summary of the repository purpose.
	Description string
	// Private controls visibility. Agent repos default to private.
	Private bool
	// DefaultBranch sets the initial branch name (default: "main").
	DefaultBranch string
	// AutoInit creates an initial commit with README if true.
	AutoInit bool
}

// CommitInfo represents metadata for a Git commit.
type CommitInfo struct {
	SHA       string
	Message   string
	Author    string
	Timestamp time.Time
	Parents   []string
}

// FileEntry represents a file in the repository tree.
type FileEntry struct {
	Path    string
	Size    int64
	Mode    string // "file", "dir", "symlink", "submodule"
	SHA     string
	IsDir   bool
}

// DiffEntry represents a single file change in a diff.
type DiffEntry struct {
	OldPath   string
	NewPath   string
	Status    string // "added", "modified", "deleted", "renamed"
	Additions int
	Deletions int
	Patch     string
}

// PullRequest represents a merge/pull request.
type PullRequest struct {
	ID          int64
	Title       string
	Body        string
	HeadBranch  string
	BaseBranch  string
	State       string // "open", "closed", "merged"
	CreatedAt   time.Time
	MergedAt    *time.Time
	MergeCommit string
}

// PROptions configures a new pull request.
type PROptions struct {
	Title      string
	Body       string
	HeadBranch string
	BaseBranch string
	Labels     []string
}

// MergeOptions configures how a pull request is merged.
type MergeOptions struct {
	// Strategy is the merge strategy: "merge", "rebase", "squash".
	Strategy string
	// CommitMessage overrides the default merge commit message.
	CommitMessage string
	// DeleteBranch removes the head branch after merge.
	DeleteBranch bool
}

// BranchInfo represents a Git branch.
type BranchInfo struct {
	Name      string
	CommitSHA string
	Protected bool
}

// CloneOptions configures a repository clone operation.
type CloneOptions struct {
	// Depth limits history depth (0 = full clone).
	Depth int
	// Branch specifies which branch to clone (default: default branch).
	Branch string
	// DestPath is the local filesystem destination.
	DestPath string
}

// AgentGitService defines the complete Git interface available to Operator-OS agents.
// All methods are context-aware and return structured errors.
type AgentGitService interface {
	// Repository Management

	// CreateRepo creates a new repository in the agent's namespace.
	CreateRepo(ctx context.Context, name string, opts RepoOptions) error

	// DeleteRepo permanently removes a repository and all its data.
	DeleteRepo(ctx context.Context, name string) error

	// ListRepos returns all repositories owned by the agent.
	ListRepos(ctx context.Context) ([]string, error)

	// Clone copies a repository to a local path.
	Clone(ctx context.Context, name string, opts CloneOptions) (string, error)

	// Branch Operations

	// CreateBranch creates a new branch from a source ref.
	CreateBranch(ctx context.Context, repo string, branch string, fromRef string) error

	// DeleteBranch removes a branch.
	DeleteBranch(ctx context.Context, repo string, branch string) error

	// ListBranches returns all branches in a repository.
	ListBranches(ctx context.Context, repo string) ([]BranchInfo, error)

	// File Operations

	// ListFiles returns the file tree at a given ref and path.
	ListFiles(ctx context.Context, repo string, ref string, path string) ([]FileEntry, error)

	// ReadFile returns the content of a file at a given ref.
	ReadFile(ctx context.Context, repo string, ref string, path string) (io.ReadCloser, error)

	// WriteFile creates or updates a file on a branch.
	// The file is staged but not committed until Commit is called.
	WriteFile(ctx context.Context, repo string, branch string, path string, content []byte) error

	// DeleteFile removes a file from a branch.
	DeleteFile(ctx context.Context, repo string, branch string, path string) error

	// Commit Operations

	// Commit creates a commit on the specified branch with all staged changes.
	Commit(ctx context.Context, repo string, branch string, message string) (*CommitInfo, error)

	// GetCommit retrieves commit metadata by SHA.
	GetCommit(ctx context.Context, repo string, sha string) (*CommitInfo, error)

	// ListCommits returns commit history for a branch.
	ListCommits(ctx context.Context, repo string, branch string, limit int) ([]CommitInfo, error)

	// Diff returns the diff between two refs.
	Diff(ctx context.Context, repo string, baseRef string, headRef string) ([]DiffEntry, error)

	// Remote Operations

	// Push pushes local changes to the remote.
	Push(ctx context.Context, repo string, branch string) error

	// Pull fetches and merges remote changes.
	Pull(ctx context.Context, repo string, branch string) error

	// Pull Request Operations

	// CreatePR creates a new pull request.
	CreatePR(ctx context.Context, repo string, opts PROptions) (*PullRequest, error)

	// GetPR retrieves a pull request by ID.
	GetPR(ctx context.Context, repo string, id int64) (*PullRequest, error)

	// ListPRs returns pull requests for a repository.
	ListPRs(ctx context.Context, repo string, state string) ([]PullRequest, error)

	// MergePR merges a pull request.
	MergePR(ctx context.Context, repo string, id int64, opts MergeOptions) (*CommitInfo, error)

	// ClosePR closes a pull request without merging.
	ClosePR(ctx context.Context, repo string, id int64) error
}

// agentGitServiceImpl is the default implementation of AgentGitService.
type agentGitServiceImpl struct {
	pool   *RepoPool
	config *Config
	hooks  *HookManager
	agentID string
}

// NewAgentGitService creates a new AgentGitService for the given agent.
func NewAgentGitService(agentID string, pool *RepoPool, config *Config, hooks *HookManager) AgentGitService {
	return &agentGitServiceImpl{
		pool:    pool,
		config:  config,
		hooks:   hooks,
		agentID: agentID,
	}
}

func (s *agentGitServiceImpl) CreateRepo(ctx context.Context, name string, opts RepoOptions) error {
	// TODO: Validate quota, create repo via Gitea internals, fire hook
	return nil
}

func (s *agentGitServiceImpl) DeleteRepo(ctx context.Context, name string) error {
	return nil
}

func (s *agentGitServiceImpl) ListRepos(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) Clone(ctx context.Context, name string, opts CloneOptions) (string, error) {
	return "", nil
}

func (s *agentGitServiceImpl) CreateBranch(ctx context.Context, repo string, branch string, fromRef string) error {
	return nil
}

func (s *agentGitServiceImpl) DeleteBranch(ctx context.Context, repo string, branch string) error {
	return nil
}

func (s *agentGitServiceImpl) ListBranches(ctx context.Context, repo string) ([]BranchInfo, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) ListFiles(ctx context.Context, repo string, ref string, path string) ([]FileEntry, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) ReadFile(ctx context.Context, repo string, ref string, path string) (io.ReadCloser, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) WriteFile(ctx context.Context, repo string, branch string, path string, content []byte) error {
	return nil
}

func (s *agentGitServiceImpl) DeleteFile(ctx context.Context, repo string, branch string, path string) error {
	return nil
}

func (s *agentGitServiceImpl) Commit(ctx context.Context, repo string, branch string, message string) (*CommitInfo, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) GetCommit(ctx context.Context, repo string, sha string) (*CommitInfo, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) ListCommits(ctx context.Context, repo string, branch string, limit int) ([]CommitInfo, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) Diff(ctx context.Context, repo string, baseRef string, headRef string) ([]DiffEntry, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) Push(ctx context.Context, repo string, branch string) error {
	return nil
}

func (s *agentGitServiceImpl) Pull(ctx context.Context, repo string, branch string) error {
	return nil
}

func (s *agentGitServiceImpl) CreatePR(ctx context.Context, repo string, opts PROptions) (*PullRequest, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) GetPR(ctx context.Context, repo string, id int64) (*PullRequest, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) ListPRs(ctx context.Context, repo string, state string) ([]PullRequest, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) MergePR(ctx context.Context, repo string, id int64, opts MergeOptions) (*CommitInfo, error) {
	return nil, nil
}

func (s *agentGitServiceImpl) ClosePR(ctx context.Context, repo string, id int64) error {
	return nil
}

// AgentClient is a convenience wrapper for external consumers that handles
// HTTP communication with an Operator Repo instance.
type AgentClient struct {
	baseURL  string
	token    string
	agentID  string
}

// NewAgentClient creates a new client for interacting with Operator Repo.
func NewAgentClient(baseURL string, token string) *AgentClient {
	return &AgentClient{
		baseURL: baseURL,
		token:   token,
	}
}
