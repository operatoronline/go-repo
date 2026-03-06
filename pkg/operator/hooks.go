// Copyright 2026 The Operator-OS Authors. MIT License.

package operator

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// EventType identifies the kind of repository event.
type EventType string

const (
	// Repository lifecycle events
	EventRepoCreated  EventType = "repo.created"
	EventRepoDeleted  EventType = "repo.deleted"

	// Branch events
	EventBranchCreated EventType = "branch.created"
	EventBranchDeleted EventType = "branch.deleted"

	// Commit events
	EventCommitPushed EventType = "commit.pushed"

	// Pull request events
	EventPRCreated EventType = "pr.created"
	EventPRMerged  EventType = "pr.merged"
	EventPRClosed  EventType = "pr.closed"

	// File events
	EventFileWritten EventType = "file.written"
	EventFileDeleted EventType = "file.deleted"

	// Namespace events
	EventNamespaceCreated  EventType = "namespace.created"
	EventNamespaceDestroyed EventType = "namespace.destroyed"

	// Quota events
	EventQuotaWarning  EventType = "quota.warning"  // 80% usage
	EventQuotaExceeded EventType = "quota.exceeded"  // 100% usage
)

// Event represents a webhook event payload.
type Event struct {
	// ID is a unique identifier for this event.
	ID string `json:"id"`
	// Type identifies the event kind.
	Type EventType `json:"type"`
	// AgentID is the agent that triggered the event.
	AgentID string `json:"agent_id"`
	// Repo is the repository name (empty for namespace events).
	Repo string `json:"repo,omitempty"`
	// Ref is the Git ref involved (branch, tag, SHA).
	Ref string `json:"ref,omitempty"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// Payload contains event-specific data.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// WebhookTarget defines a destination for event delivery.
type WebhookTarget struct {
	// URL is the HTTP endpoint to POST events to.
	URL string
	// Secret is the HMAC-SHA256 signing key for this target.
	Secret string
	// Events filters which event types are delivered. Empty means all events.
	Events []EventType
	// Active enables/disables this target.
	Active bool
	// RetryCount is the number of delivery retries on failure.
	RetryCount int
	// TimeoutSeconds is the HTTP request timeout.
	TimeoutSeconds int
}

// DeliveryResult records the outcome of a webhook delivery attempt.
type DeliveryResult struct {
	TargetURL    string
	EventID      string
	StatusCode   int
	Success      bool
	Error        string
	Duration     time.Duration
	Attempts     int
	DeliveredAt  time.Time
}

// EventHandler is a callback function for processing events locally.
type EventHandler func(ctx context.Context, event *Event) error

// HookManager coordinates event emission, webhook delivery, and local event handlers.
type HookManager struct {
	config   *Config
	client   *http.Client
	mu       sync.RWMutex
	targets  map[string]*WebhookTarget // keyed by URL
	handlers map[EventType][]EventHandler
	history  []DeliveryResult
}

// NewHookManager creates a new webhook/event manager.
func NewHookManager(config *Config) *HookManager {
	return &HookManager{
		config: config,
		client: &http.Client{
			Timeout: config.WebhookTimeout,
		},
		targets:  make(map[string]*WebhookTarget),
		handlers: make(map[EventType][]EventHandler),
	}
}

// RegisterTarget adds a webhook delivery target.
func (h *HookManager) RegisterTarget(target *WebhookTarget) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if target.URL == "" {
		return fmt.Errorf("webhook target URL is required")
	}

	if target.TimeoutSeconds == 0 {
		target.TimeoutSeconds = 10
	}
	if target.RetryCount == 0 {
		target.RetryCount = 3
	}

	h.targets[target.URL] = target
	return nil
}

// RemoveTarget removes a webhook delivery target by URL.
func (h *HookManager) RemoveTarget(url string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.targets, url)
}

// ListTargets returns all registered webhook targets.
func (h *HookManager) ListTargets() []*WebhookTarget {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*WebhookTarget, 0, len(h.targets))
	for _, t := range h.targets {
		result = append(result, t)
	}
	return result
}

// OnEvent registers a local event handler for a specific event type.
func (h *HookManager) OnEvent(eventType EventType, handler EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[eventType] = append(h.handlers[eventType], handler)
}

// Emit fires an event, delivering it to all matching webhook targets and local handlers.
func (h *HookManager) Emit(ctx context.Context, event *Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Dispatch to local handlers
	h.mu.RLock()
	handlers := h.handlers[event.Type]
	targets := make([]*WebhookTarget, 0)
	for _, t := range h.targets {
		if t.Active && h.shouldDeliver(t, event.Type) {
			targets = append(targets, t)
		}
	}
	h.mu.RUnlock()

	// Run local handlers synchronously
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			// Log but don't fail on handler errors
			_ = err
		}
	}

	// Deliver to webhook targets asynchronously
	for _, target := range targets {
		go h.deliver(ctx, target, event)
	}

	return nil
}

// shouldDeliver checks if a target wants a specific event type.
func (h *HookManager) shouldDeliver(target *WebhookTarget, eventType EventType) bool {
	if len(target.Events) == 0 {
		return true // empty filter = all events
	}
	for _, t := range target.Events {
		if t == eventType {
			return true
		}
	}
	return false
}

// deliver sends an event to a webhook target with retries.
func (h *HookManager) deliver(ctx context.Context, target *WebhookTarget, event *Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	var lastErr error
	for attempt := 0; attempt <= target.RetryCount; attempt++ {
		start := time.Now()

		req, err := http.NewRequestWithContext(ctx, "POST", target.URL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Operator-Event", string(event.Type))
		req.Header.Set("X-Operator-Delivery", event.ID)

		// Sign payload with HMAC-SHA256
		if target.Secret != "" {
			sig := signPayload(payload, target.Secret)
			req.Header.Set("X-Operator-Signature", "sha256="+sig)
		}

		resp, err := h.client.Do(req)
		duration := time.Since(start)

		result := DeliveryResult{
			TargetURL:   target.URL,
			EventID:     event.ID,
			Duration:    duration,
			Attempts:    attempt + 1,
			DeliveredAt: time.Now(),
		}

		if err != nil {
			result.Error = err.Error()
			lastErr = err
		} else {
			result.StatusCode = resp.StatusCode
			result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
			resp.Body.Close()

			if result.Success {
				h.recordResult(result)
				return
			}
			lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		}

		h.recordResult(result)

		// Exponential backoff between retries
		if attempt < target.RetryCount {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}

	_ = lastErr
}

// recordResult stores a delivery result in the history ring buffer.
func (h *HookManager) recordResult(result DeliveryResult) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.history = append(h.history, result)
	// Keep last 1000 results
	if len(h.history) > 1000 {
		h.history = h.history[len(h.history)-1000:]
	}
}

// DeliveryHistory returns recent webhook delivery results.
func (h *HookManager) DeliveryHistory(limit int) []DeliveryResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.history) {
		limit = len(h.history)
	}
	start := len(h.history) - limit
	result := make([]DeliveryResult, limit)
	copy(result, h.history[start:])
	return result
}

// signPayload computes HMAC-SHA256 signature for webhook authentication.
func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
