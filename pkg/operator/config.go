// Copyright 2026 The Operator-OS Authors. MIT License.

package operator

import (
	"os"
	"strconv"
	"time"
)

// TierLimits defines resource constraints for an agent tier.
type TierLimits struct {
	// MaxRepos is the maximum number of repositories an agent can create.
	MaxRepos int
	// StorageLimit is the maximum total storage in bytes.
	StorageLimit int64
	// MaxFileSize is the maximum size of a single file in bytes.
	MaxFileSize int64
	// MaxBranches is the maximum number of branches per repository.
	MaxBranches int
	// RetentionDays is how long inactive repos are kept before cleanup.
	RetentionDays int
}

// Predefined tier configurations.
var (
	// TierFree is for evaluation and testing agents.
	TierFree = TierLimits{
		MaxRepos:      5,
		StorageLimit:  100 * 1024 * 1024, // 100 MB
		MaxFileSize:   10 * 1024 * 1024,  // 10 MB
		MaxBranches:   10,
		RetentionDays: 30,
	}

	// TierStandard is for production agent workloads.
	TierStandard = TierLimits{
		MaxRepos:      50,
		StorageLimit:  1024 * 1024 * 1024, // 1 GB
		MaxFileSize:   50 * 1024 * 1024,   // 50 MB
		MaxBranches:   50,
		RetentionDays: 90,
	}

	// TierPro is for high-throughput agent clusters.
	TierPro = TierLimits{
		MaxRepos:      500,
		StorageLimit:  10 * 1024 * 1024 * 1024, // 10 GB
		MaxFileSize:   100 * 1024 * 1024,        // 100 MB
		MaxBranches:   200,
		RetentionDays: 365,
	}

	// TierUnlimited removes all restrictions (self-hosted).
	TierUnlimited = TierLimits{
		MaxRepos:      0, // 0 = unlimited
		StorageLimit:  0,
		MaxFileSize:   0,
		MaxBranches:   0,
		RetentionDays: 0, // 0 = never cleanup
	}
)

// Config holds all Operator-OS specific configuration for go-repo.
type Config struct {
	// DataDir is the root directory for all repository data.
	DataDir string

	// ListenAddr is the address the server binds to.
	ListenAddr string

	// ListenPort is the port the server binds to.
	ListenPort int

	// DatabasePath is the path to the embedded SQLite database.
	DatabasePath string

	// DefaultTier is the tier assigned to new agents.
	DefaultTier string

	// TierOverrides maps agent IDs to specific tier names.
	TierOverrides map[string]string

	// WebhookSecret is the shared secret for signing webhook payloads.
	WebhookSecret string

	// WebhookTimeout is the maximum duration for webhook delivery.
	WebhookTimeout time.Duration

	// CleanupInterval is how often the cleanup job runs.
	CleanupInterval time.Duration

	// EnableAPI enables the Gitea REST API (default: true).
	EnableAPI bool

	// EnableWebUI enables the Gitea web interface (default: false for agent-only mode).
	EnableWebUI bool

	// EnableSSH enables Git SSH access (default: false, agents use HTTP).
	EnableSSH bool

	// LogLevel sets the logging verbosity: "debug", "info", "warn", "error".
	LogLevel string
}

// DefaultConfig returns a Config with sensible defaults for Operator-OS.
func DefaultConfig() *Config {
	return &Config{
		DataDir:         envOrDefault("OPERATOR_DATA_DIR", "/data"),
		ListenAddr:      envOrDefault("OPERATOR_LISTEN_ADDR", "0.0.0.0"),
		ListenPort:      envOrDefaultInt("OPERATOR_LISTEN_PORT", 3000),
		DatabasePath:    envOrDefault("OPERATOR_DB_PATH", "/data/operator-repo.db"),
		DefaultTier:     envOrDefault("OPERATOR_DEFAULT_TIER", "standard"),
		TierOverrides:   make(map[string]string),
		WebhookSecret:   envOrDefault("OPERATOR_WEBHOOK_SECRET", ""),
		WebhookTimeout:  10 * time.Second,
		CleanupInterval: 24 * time.Hour,
		EnableAPI:       true,
		EnableWebUI:     envOrDefaultBool("OPERATOR_ENABLE_WEB_UI", false),
		EnableSSH:       envOrDefaultBool("OPERATOR_ENABLE_SSH", false),
		LogLevel:        envOrDefault("OPERATOR_LOG_LEVEL", "info"),
	}
}

// GetAgentTier returns the tier name for a given agent ID.
// Returns the override tier if set, otherwise the default tier.
func (c *Config) GetAgentTier(agentID string) string {
	if tier, ok := c.TierOverrides[agentID]; ok {
		return tier
	}
	return c.DefaultTier
}

// GetTierLimits returns the TierLimits for a named tier.
func (c *Config) GetTierLimits(tier string) TierLimits {
	switch tier {
	case "free":
		return TierFree
	case "standard":
		return TierStandard
	case "pro":
		return TierPro
	case "unlimited":
		return TierUnlimited
	default:
		return TierStandard
	}
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	// TODO: Check that DataDir exists or can be created
	// TODO: Validate port range
	// TODO: Ensure DatabasePath parent directory exists
	return nil
}

// helper functions for environment variable parsing

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envOrDefaultBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
