// Package service projects the upgrade worker state for the management
// console and validates every operator request before it reaches the worker.
package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	operationsdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/operations/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/repository"
	"github.com/google/uuid"
)

// Service owns the read-mostly service inventory and its upgrade commands.
type Service struct {
	repository repository.Repository
	catalog    []domain.CatalogEntry
	clock      func() time.Time
}

// Options contains service version dependencies.
type Options struct {
	Repository repository.Repository
	Clock      func() time.Time
}

// UpgradePlan is a validated batch of service upgrades.
type UpgradePlan struct {
	TargetService string
	TargetVersion string
	UpgradesSelf  bool
}

// New creates the service version service.
func New(options Options) (*Service, error) {
	if options.Repository == nil {
		return nil, errors.New("service version repository is required")
	}
	timeSource := options.Clock
	if timeSource == nil {
		timeSource = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		repository: options.Repository,
		catalog:    defaultCatalog(),
		clock:      timeSource,
	}, nil
}

// Snapshot merges the worker snapshot with the catalogue so every managed
// service is visible even before the worker has reported on it.
func (s *Service) Snapshot(ctx context.Context) ([]domain.Service, time.Time, error) {
	snapshot, err := s.repository.Status(ctx)
	if err != nil && !errors.Is(err, domain.ErrStateUnavailable) {
		return nil, time.Time{}, err
	}
	reported := make(map[string]domain.Service, len(snapshotServices(snapshot)))
	for _, service := range snapshotServices(snapshot) {
		reported[service.ID] = service
	}
	services := make([]domain.Service, 0, len(s.catalog))
	for _, entry := range s.catalog {
		service := reported[entry.ID]
		service.ID = entry.ID
		service.DisplayName = entry.DisplayName
		service.Role = entry.Role
		service.Image = entry.Image
		service.IsSelf = entry.IsSelf
		if service.Status == "" {
			service.Status = domain.ServiceStatusUnknown
		}
		// An infrastructure image is detected but never auto-upgraded, so the
		// console must not offer an action the worker would refuse.
		service.CanUpgrade = entry.AutoUpgrade &&
			service.Status == domain.ServiceStatusOutdated &&
			service.LatestVersion != "" &&
			CompareVersions(service.LatestVersion, service.CurrentVersion) > 0
		services = append(services, service)
	}
	sort.SliceStable(services, func(left int, right int) bool {
		return catalogRank(s.catalog, services[left].ID) <
			catalogRank(s.catalog, services[right].ID)
	})
	var checkedAt time.Time
	if snapshot != nil {
		checkedAt = snapshot.GeneratedAt
	}
	return services, checkedAt, nil
}

// AllCurrent reports whether every service is at its newest version. Unknown
// services count as not-current so the console never claims success blindly.
func (s *Service) AllCurrent(services []domain.Service) bool {
	for _, service := range services {
		if service.Status != domain.ServiceStatusCurrent {
			return false
		}
	}
	return len(services) > 0
}

// RequestCheck asks the worker to refresh its snapshot immediately.
func (s *Service) RequestCheck(ctx context.Context) error {
	return s.repository.RequestCheck(ctx)
}

// PlanUpgrade validates a single-service upgrade request against the latest
// snapshot and returns the concrete target version to enqueue.
func (s *Service) PlanUpgrade(
	ctx context.Context,
	serviceID string,
) (UpgradePlan, error) {
	services, _, err := s.Snapshot(ctx)
	if err != nil {
		return UpgradePlan{}, err
	}
	serviceID = strings.TrimSpace(serviceID)
	for _, service := range services {
		if service.ID != serviceID {
			continue
		}
		if !service.CanUpgrade {
			return UpgradePlan{}, domain.ErrServiceNotUpgradable
		}
		return UpgradePlan{
			TargetService: service.ID,
			TargetVersion: service.LatestVersion,
			UpgradesSelf:  service.IsSelf,
		}, nil
	}
	return UpgradePlan{}, domain.ErrServiceNotFound
}

// PlanUpgradeAll validates the batch upgrade in the worker's required order.
func (s *Service) PlanUpgradeAll(
	ctx context.Context,
) ([]UpgradePlan, error) {
	services, _, err := s.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]domain.Service, len(services))
	for _, service := range services {
		byID[service.ID] = service
	}
	plans := make([]UpgradePlan, 0, len(services))
	for _, entry := range s.catalog {
		if !entry.AutoUpgrade {
			continue
		}
		service, exists := byID[entry.ID]
		if !exists || !service.CanUpgrade {
			continue
		}
		plans = append(plans, UpgradePlan{
			TargetService: service.ID,
			TargetVersion: service.LatestVersion,
			UpgradesSelf:  service.IsSelf,
		})
	}
	if len(plans) == 0 {
		return nil, domain.ErrNoUpgradeAvailable
	}
	return plans, nil
}

// Enqueue persists one upgrade command for the worker.
func (s *Service) Enqueue(
	ctx context.Context,
	plan UpgradePlan,
	actorAccountID string,
) (*domain.Operation, error) {
	if err := s.ensureIdle(ctx); err != nil {
		return nil, err
	}
	operationID := uuid.NewString()
	now := s.clock().UTC()
	request := &domain.UpgradeRequest{
		ID:            operationID,
		TargetService: plan.TargetService,
		TargetVersion: plan.TargetVersion,
		RequestedAt:   now,
		RequestedBy:   strings.TrimSpace(actorAccountID),
	}
	if err := s.repository.Enqueue(ctx, request); err != nil {
		return nil, err
	}
	return &domain.Operation{
		ID:            operationID,
		TargetService: plan.TargetService,
		TargetVersion: plan.TargetVersion,
		Status:        domain.OperationStatusQueued,
		Message:       "升级任务已排队，正在准备更新",
		RequestedAt:   &now,
	}, nil
}

// ListOperations returns the upgrade history newest first.
func (s *Service) ListOperations(ctx context.Context) ([]domain.Operation, error) {
	return s.repository.ListOperations(ctx)
}

// EnqueueAll queues one command per outdated service, in the worker's fixed
// order. The worker still runs them one at a time; queueing them together only
// removes the need for an operator to click each service.
func (s *Service) EnqueueAll(
	ctx context.Context,
	plans []UpgradePlan,
	actorAccountID string,
) ([]domain.Operation, error) {
	if len(plans) == 0 {
		return nil, domain.ErrNoUpgradeAvailable
	}
	if err := s.ensureIdle(ctx); err != nil {
		return nil, err
	}
	actorAccountID = strings.TrimSpace(actorAccountID)
	operations := make([]domain.Operation, 0, len(plans))
	for _, plan := range plans {
		operationID := uuid.NewString()
		now := s.clock().UTC()
		request := &domain.UpgradeRequest{
			ID:            operationID,
			TargetService: plan.TargetService,
			TargetVersion: plan.TargetVersion,
			RequestedAt:   now,
			RequestedBy:   actorAccountID,
		}
		if err := s.repository.Enqueue(ctx, request); err != nil {
			return operations, err
		}
		operations = append(operations, domain.Operation{
			ID:            operationID,
			TargetService: plan.TargetService,
			TargetVersion: plan.TargetVersion,
			Status:        domain.OperationStatusQueued,
			Message:       "升级任务已排队，正在准备更新",
			RequestedAt:   &now,
		})
	}
	return operations, nil
}

// FindOperation returns one upgrade task by id.
func (s *Service) FindOperation(
	ctx context.Context,
	operationID string,
) (*domain.Operation, error) {
	return s.repository.FindOperation(ctx, operationID)
}

// ensureIdle refuses a new upgrade while another one is still active so the
// worker's single-upgrade lock is never overloaded.
func (s *Service) ensureIdle(ctx context.Context) error {
	operations, err := s.repository.ListOperations(ctx)
	if err != nil {
		return err
	}
	for _, operation := range operations {
		if domain.IsActiveOperation(operation.Status) {
			return domain.ErrUpgradeInProgress
		}
	}
	return nil
}

func snapshotServices(snapshot *domain.StatusSnapshot) []domain.Service {
	if snapshot == nil {
		return nil
	}
	return snapshot.Services
}

// catalogRank preserves the operator-facing order from defaultCatalog.
func catalogRank(catalog []domain.CatalogEntry, id string) int {
	for index, entry := range catalog {
		if entry.ID == id {
			return index
		}
	}
	return len(catalog)
}

// defaultCatalog is the single source of truth for managed image identity and
// upgrade eligibility. Infrastructure images are observation-only.
func defaultCatalog() []domain.CatalogEntry {
	return []domain.CatalogEntry{
		{
			ID:             "sub2api",
			DisplayName:    "AI 网关",
			Role:           "AI 网关",
			Image:          "ghcr.io/tissyboxc/sub2api",
			AutoUpgrade:    true,
			PlatformScoped: false,
		},
		{
			ID:             "device_platform",
			DisplayName:    "设备平台",
			Role:           "平台服务",
			Image:          "ghcr.io/tissyboxc/sprout-device-platform",
			AutoUpgrade:    true,
			PlatformScoped: true,
		},
		{
			ID:             "voice_gateway",
			DisplayName:    "语音网关",
			Role:           "平台服务",
			Image:          "ghcr.io/tissyboxc/sprout-voice-gateway",
			AutoUpgrade:    true,
			PlatformScoped: true,
		},
		{
			ID:             "admin_web",
			DisplayName:    "品牌管理端",
			Role:           "平台服务",
			Image:          "ghcr.io/tissyboxc/sprout-admin-web",
			AutoUpgrade:    true,
			PlatformScoped: true,
			IsSelf:         true,
		},
		{
			ID:          "postgres",
			DisplayName: "数据库",
			Role:        "基础设施",
			Image:       "postgres",
		},
		{
			ID:          "redis",
			DisplayName: "缓存服务",
			Role:        "基础设施",
			Image:       "redis",
		},
		{
			ID:          "mqtt",
			DisplayName: "消息服务",
			Role:        "基础设施",
			Image:       "eclipse-mosquitto",
		},
		{
			ID:          "download_init",
			DisplayName: "下载服务初始化",
			Role:        "发布服务",
			Image:       "alpine",
		},
		{
			ID:          "download_ftp",
			DisplayName: "发布文件传输",
			Role:        "发布服务",
			Image:       "atmoz/sftp",
		},
		{
			ID:          "download_http",
			DisplayName: "发布文件下载",
			Role:        "发布服务",
			Image:       "nginx",
		},
	}
}

// CompareVersions compares dotted semantic versions numerically.
func CompareVersions(left string, right string) int {
	return operationsdomain.CompareVersions(left, right)
}
