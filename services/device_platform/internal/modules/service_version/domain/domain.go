// Package domain defines service version and upgrade command types.
//
// The device platform never talks to the Docker daemon. It only reads the
// status snapshot and writes command files that the isolated upgrade worker
// consumes, so a compromised public API cannot control the host.
package domain

import (
	"errors"
	"time"

	operationsdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/operations/domain"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceNotUpgradable = errors.New("service cannot be upgraded")
	ErrUpgradeInProgress    = errors.New("an upgrade is already in progress")
	ErrOperationNotFound    = errors.New("upgrade operation not found")
	ErrStateUnavailable     = errors.New("service version state unavailable")
	ErrNoUpgradeAvailable   = errors.New("no service needs an upgrade")
)

// Service status values describe a managed container image relative to the
// newest version published in its repository.
const (
	ServiceStatusCurrent  = "current"
	ServiceStatusOutdated = "outdated"
	ServiceStatusUnknown  = "unknown"
	ServiceStatusUpdating = "updating"
	ServiceStatusFailed   = "failed"
)

// Upgrade operation status values mirror the worker's result files.
const (
	OperationStatusQueued     = "queued"
	OperationStatusRunning    = "running"
	OperationStatusRecovering = "recovering"
	OperationStatusSucceeded  = "succeeded"
	OperationStatusFailed     = "failed"
)

// Service is one managed container image as projected for the console.
//
// CurrentVersion, LatestVersion, and ReleaseURL are advisory: when the worker
// cannot confirm a value it leaves the field empty and sets Status to unknown
// instead of inventing a version.
type Service struct {
	ID             string     `json:"id"`
	DisplayName    string     `json:"display_name"`
	Role           string     `json:"role"`
	Image          string     `json:"image"`
	CurrentVersion string     `json:"current_version"`
	LatestVersion  string     `json:"latest_version"`
	Status         string     `json:"status"`
	ReleaseURL     string     `json:"release_url"`
	IsSelf         bool       `json:"is_self"`
	CanUpgrade     bool       `json:"can_upgrade"`
	LastCheckedAt  *time.Time `json:"last_checked_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// Operation is one queued, running, or finished upgrade task.
type Operation struct {
	ID             string     `json:"id"`
	TargetService  string     `json:"target_service"`
	CurrentVersion string     `json:"current_version"`
	TargetVersion  string     `json:"target_version"`
	Status         string     `json:"status"`
	Message        string     `json:"message"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	LogTail        string     `json:"log_tail"`
	RequestedBy    string     `json:"requested_by,omitempty"`
	RequestedAt    *time.Time `json:"requested_at,omitempty"`
}

// StatusSnapshot is the worker-produced view of every managed image.
type StatusSnapshot struct {
	GeneratedAt time.Time `json:"generated_at"`
	Services    []Service `json:"services"`
}

// CheckRequest asks the worker to refresh the status snapshot immediately.
type CheckRequest struct {
	RequestedAt time.Time `json:"requested_at"`
}

// UpgradeRequest is the worker's queue file for one service upgrade.
type UpgradeRequest struct {
	ID            string    `json:"id"`
	TargetService string    `json:"target_service"`
	TargetVersion string    `json:"target_version"`
	RequestedAt   time.Time `json:"requested_at"`
	RequestedBy   string    `json:"requested_by"`
}

// CatalogEntry is the operator-facing identity of a managed image.
//
// AutoUpgrade is false for infrastructure images: operators can see that a
// newer tag exists but the worker never restarts Postgres, Redis, or the
// download services on its own.
type CatalogEntry struct {
	ID             string
	DisplayName    string
	Role           string
	Image          string
	AutoUpgrade    bool
	PlatformScoped bool
	IsSelf         bool
}

// CompareVersions compares dotted semantic versions numerically.
func CompareVersions(left string, right string) int {
	return operationsdomain.CompareVersions(left, right)
}

// IsActiveOperation reports whether an operation still owns the upgrade lock.
func IsActiveOperation(status string) bool {
	switch status {
	case OperationStatusQueued,
		OperationStatusRunning,
		OperationStatusRecovering:
		return true
	default:
		return false
	}
}
