// Copyright 2026 The Operator-OS Authors. MIT License.

package operator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AgentNamespace represents an isolated namespace for a single agent's repositories.
type AgentNamespace struct {
	// AgentID is the unique identifier for the agent.
	AgentID string
	// RepoCount is the current number of repositories.
	RepoCount int
	// StorageUsed is the total storage consumed in bytes.
	StorageUsed int64
	// CreatedAt is when the namespace was provisioned.
	CreatedAt time.Time
	// LastActive is the most recent operation timestamp.
	LastActive time.Time
}

// QuotaStatus reports an agent's resource consumption against limits.
type QuotaStatus struct {
	AgentID       string
	Tier          string
	RepoCount     int
	RepoLimit     int
	StorageUsed   int64
	StorageLimit  int64
	IsOverQuota   bool
}

// CleanupPolicy defines when and how inactive resources are reclaimed.
type CleanupPolicy struct {
	// MaxInactiveDays is the number of days after which idle repos may be archived.
	MaxInactiveDays int
	// ArchiveBeforeDelete requires archival before permanent deletion.
	ArchiveBeforeDelete bool
	// DryRun logs actions without executing them.
	DryRun bool
}

// CleanupResult summarizes a cleanup operation.
type CleanupResult struct {
	NamespacesScanned  int
	ReposArchived      int
	ReposDeleted       int
	StorageReclaimed   int64
	Errors             []error
}

// RepoPool manages multi-tenant repository isolation, quota enforcement,
// and lifecycle operations across all agent namespaces.
type RepoPool struct {
	config     *Config
	hooks      *HookManager
	mu         sync.RWMutex
	namespaces map[string]*AgentNamespace
}

// NewRepoPool creates a new multi-tenant repository pool.
func NewRepoPool(config *Config, hooks *HookManager) *RepoPool {
	return &RepoPool{
		config:     config,
		hooks:      hooks,
		namespaces: make(map[string]*AgentNamespace),
	}
}

// ProvisionNamespace creates an isolated namespace for a new agent.
// Returns an error if the agent already has a namespace.
func (p *RepoPool) ProvisionNamespace(ctx context.Context, agentID string, tier string) (*AgentNamespace, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.namespaces[agentID]; exists {
		return nil, fmt.Errorf("namespace already exists for agent %s", agentID)
	}

	ns := &AgentNamespace{
		AgentID:   agentID,
		CreatedAt: time.Now(),
		LastActive: time.Now(),
	}
	p.namespaces[agentID] = ns

	// TODO: Create Gitea user/org for namespace, set up filesystem paths
	// TODO: Fire NamespaceCreated hook

	return ns, nil
}

// DestroyNamespace removes an agent's namespace and all contained repositories.
// This is irreversible.
func (p *RepoPool) DestroyNamespace(ctx context.Context, agentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.namespaces[agentID]; !exists {
		return fmt.Errorf("namespace not found for agent %s", agentID)
	}

	// TODO: Delete all repos, remove Gitea user, clean filesystem
	// TODO: Fire NamespaceDestroyed hook

	delete(p.namespaces, agentID)
	return nil
}

// GetNamespace returns the namespace for an agent.
func (p *RepoPool) GetNamespace(ctx context.Context, agentID string) (*AgentNamespace, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ns, exists := p.namespaces[agentID]
	if !exists {
		return nil, fmt.Errorf("namespace not found for agent %s", agentID)
	}
	return ns, nil
}

// ListNamespaces returns all active agent namespaces.
func (p *RepoPool) ListNamespaces(ctx context.Context) ([]*AgentNamespace, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*AgentNamespace, 0, len(p.namespaces))
	for _, ns := range p.namespaces {
		result = append(result, ns)
	}
	return result, nil
}

// CheckQuota returns the current quota status for an agent.
func (p *RepoPool) CheckQuota(ctx context.Context, agentID string) (*QuotaStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ns, exists := p.namespaces[agentID]
	if !exists {
		return nil, fmt.Errorf("namespace not found for agent %s", agentID)
	}

	tier := p.config.GetAgentTier(agentID)
	limits := p.config.GetTierLimits(tier)

	return &QuotaStatus{
		AgentID:      agentID,
		Tier:         tier,
		RepoCount:    ns.RepoCount,
		RepoLimit:    limits.MaxRepos,
		StorageUsed:  ns.StorageUsed,
		StorageLimit: limits.StorageLimit,
		IsOverQuota:  ns.StorageUsed > limits.StorageLimit || ns.RepoCount > limits.MaxRepos,
	}, nil
}

// EnforceQuota checks if an agent can perform a storage-consuming operation.
// Returns nil if within quota, or an error describing the exceeded limit.
func (p *RepoPool) EnforceQuota(ctx context.Context, agentID string, additionalBytes int64) error {
	status, err := p.CheckQuota(ctx, agentID)
	if err != nil {
		return err
	}

	tier := p.config.GetAgentTier(agentID)
	limits := p.config.GetTierLimits(tier)

	if status.StorageUsed+additionalBytes > limits.StorageLimit {
		return fmt.Errorf("storage quota exceeded for agent %s (tier: %s, used: %d, limit: %d)",
			agentID, tier, status.StorageUsed+additionalBytes, limits.StorageLimit)
	}

	return nil
}

// RunCleanup performs garbage collection across all namespaces according to policy.
func (p *RepoPool) RunCleanup(ctx context.Context, policy CleanupPolicy) (*CleanupResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := &CleanupResult{}
	cutoff := time.Now().AddDate(0, 0, -policy.MaxInactiveDays)

	for _, ns := range p.namespaces {
		result.NamespacesScanned++

		if ns.LastActive.Before(cutoff) {
			// TODO: Archive or delete inactive repos
			// TODO: Calculate storage reclaimed
			if !policy.DryRun {
				// Execute cleanup
			}
		}
	}

	return result, nil
}

// UpdateStorageUsage recalculates storage consumption for an agent namespace.
func (p *RepoPool) UpdateStorageUsage(ctx context.Context, agentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ns, exists := p.namespaces[agentID]
	if !exists {
		return fmt.Errorf("namespace not found for agent %s", agentID)
	}

	// TODO: Walk filesystem and calculate actual storage usage
	_ = ns
	return nil
}

// GetServiceForAgent returns an AgentGitService bound to the specified agent's namespace.
func (p *RepoPool) GetServiceForAgent(agentID string) AgentGitService {
	return NewAgentGitService(agentID, p, p.config, p.hooks)
}
