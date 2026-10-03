package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/repository"
)

func newTestService(t *testing.T, snapshotJSON string) (*Service, string) {
	t.Helper()
	stateDir := t.TempDir()
	if snapshotJSON != "" {
		if err := os.WriteFile(
			filepath.Join(stateDir, "status.json"),
			[]byte(snapshotJSON),
			0o600,
		); err != nil {
			t.Fatalf("write snapshot: %v", err)
		}
	}
	service, err := New(Options{
		Repository: repository.NewFileRepository(stateDir),
		Clock:      func() time.Time { return time.Unix(0, 0).UTC() },
	})
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	return service, stateDir
}

func TestSnapshotMergesCatalogAndMarksUpgradable(t *testing.T) {
	service, _ := newTestService(t, `{
		"generated_at": "2026-10-03T00:00:00Z",
		"services": [
			{"id":"device_platform","current_version":"0.9.0","latest_version":"0.10.0","status":"outdated","release_url":"https://example.test/v0.10.0"}
		]
	}`)

	services, checkedAt, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() returned unexpected error: %v", err)
	}
	if checkedAt.IsZero() {
		t.Fatal("expected a non-zero checked timestamp")
	}
	var platform *domain.Service
	for index := range services {
		if services[index].ID == "device_platform" {
			platform = &services[index]
		}
	}
	if platform == nil {
		t.Fatal("expected device_platform in the snapshot")
	}
	if !platform.CanUpgrade {
		t.Fatal("expected an outdated platform to be upgradable")
	}
	if platform.DisplayName == "" {
		t.Fatal("expected the catalogue to supply a display name")
	}
}

func TestInfrastructureNeverUpgradable(t *testing.T) {
	service, _ := newTestService(t, `{
		"generated_at": "2026-10-03T00:00:00Z",
		"services": [
			{"id":"postgres","current_version":"15","latest_version":"16","status":"outdated"}
		]
	}`)

	services, _, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() returned unexpected error: %v", err)
	}
	for _, item := range services {
		if item.ID == "postgres" && item.CanUpgrade {
			t.Fatal("infrastructure images must not be auto-upgraded")
		}
	}
}

func TestSnapshotWithoutWorkerStateIsUnknown(t *testing.T) {
	service, _ := newTestService(t, "")

	services, _, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() returned unexpected error: %v", err)
	}
	if len(services) == 0 {
		t.Fatal("expected the catalogue to be returned even without a snapshot")
	}
	for _, item := range services {
		if item.Status != domain.ServiceStatusUnknown {
			t.Fatalf("expected unknown status, got %q for %s", item.Status, item.ID)
		}
		if item.CanUpgrade {
			t.Fatalf("an unknown service must not be upgradable: %s", item.ID)
		}
	}
}

func TestPlanUpgradeRejectsUnknownService(t *testing.T) {
	service, _ := newTestService(t, `{"generated_at":"2026-10-03T00:00:00Z","services":[]}`)

	_, err := service.PlanUpgrade(context.Background(), "does_not_exist")
	if !errors.Is(err, domain.ErrServiceNotFound) {
		t.Fatalf("expected ErrServiceNotFound, got %v", err)
	}
}

func TestPlanUpgradeRejectsCurrentService(t *testing.T) {
	service, _ := newTestService(t, `{
		"generated_at":"2026-10-03T00:00:00Z",
		"services":[{"id":"device_platform","current_version":"0.10.0","latest_version":"0.10.0","status":"current"}]
	}`)

	_, err := service.PlanUpgrade(context.Background(), "device_platform")
	if !errors.Is(err, domain.ErrServiceNotUpgradable) {
		t.Fatalf("expected ErrServiceNotUpgradable, got %v", err)
	}
}

func TestEnqueueAndListOperations(t *testing.T) {
	service, stateDir := newTestService(t, `{
		"generated_at":"2026-10-03T00:00:00Z",
		"services":[{"id":"device_platform","current_version":"0.9.0","latest_version":"0.10.0","status":"outdated"}]
	}`)

	plan, err := service.PlanUpgrade(context.Background(), "device_platform")
	if err != nil {
		t.Fatalf("PlanUpgrade() returned unexpected error: %v", err)
	}
	operation, err := service.Enqueue(context.Background(), plan, "admin-1")
	if err != nil {
		t.Fatalf("Enqueue() returned unexpected error: %v", err)
	}
	if operation.Status != domain.OperationStatusQueued {
		t.Fatalf("expected queued operation, got %q", operation.Status)
	}
	if _, err := os.Stat(
		filepath.Join(stateDir, "queue", operation.ID+".json"),
	); err != nil {
		t.Fatalf("expected queue file: %v", err)
	}

	// A second enqueue while the worker result is absent still counts the
	// first as inactive, so a fresh upgrade is allowed. The lock is enforced
	// again once the worker writes an active result file.
	if _, err := service.Enqueue(context.Background(), plan, "admin-1"); err != nil {
		t.Fatalf("Enqueue() returned unexpected error: %v", err)
	}
}

func TestEnqueueBlockedByActiveOperation(t *testing.T) {
	service, stateDir := newTestService(t, `{
		"generated_at":"2026-10-03T00:00:00Z",
		"services":[{"id":"device_platform","current_version":"0.9.0","latest_version":"0.10.0","status":"outdated"}]
	}`)
	resultsDir := filepath.Join(stateDir, "results")
	if err := os.MkdirAll(resultsDir, 0o750); err != nil {
		t.Fatalf("create results dir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(resultsDir, "active.json"),
		[]byte(`{"id":"active","target_service":"voice_gateway","status":"running"}`),
		0o600,
	); err != nil {
		t.Fatalf("write active result: %v", err)
	}

	plan, err := service.PlanUpgrade(context.Background(), "device_platform")
	if err != nil {
		t.Fatalf("PlanUpgrade() returned unexpected error: %v", err)
	}
	if _, err := service.Enqueue(context.Background(), plan, "admin-1"); !errors.Is(
		err,
		domain.ErrUpgradeInProgress,
	) {
		t.Fatalf("expected ErrUpgradeInProgress, got %v", err)
	}
}

func TestPlanUpgradeAllOrdersPlatformBeforeSelf(t *testing.T) {
	service, _ := newTestService(t, `{
		"generated_at":"2026-10-03T00:00:00Z",
		"services":[
			{"id":"admin_web","current_version":"0.9.0","latest_version":"0.10.0","status":"outdated"},
			{"id":"device_platform","current_version":"0.9.0","latest_version":"0.10.0","status":"outdated"},
			{"id":"sub2api","current_version":"0.2.15","latest_version":"0.2.16","status":"outdated"}
		]
	}`)

	plans, err := service.PlanUpgradeAll(context.Background())
	if err != nil {
		t.Fatalf("PlanUpgradeAll() returned unexpected error: %v", err)
	}
	if len(plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(plans))
	}
	order := []string{plans[0].TargetService, plans[1].TargetService, plans[2].TargetService}
	want := []string{"sub2api", "device_platform", "admin_web"}
	for index := range want {
		if order[index] != want[index] {
			t.Fatalf("unexpected upgrade order: %v", order)
		}
	}
}
