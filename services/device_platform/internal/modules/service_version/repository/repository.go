// Package repository reads and writes the upgrade worker's shared state.
//
// The worker owns the Docker socket; this repository only performs atomic
// file operations inside a single state directory. Writes go to a temporary
// file in the same directory and are renamed into place so the worker never
// reads a half-written command.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/domain"
)

// Repository exposes the worker state files used by the management API.
type Repository interface {
	Status(ctx context.Context) (*domain.StatusSnapshot, error)
	RequestCheck(ctx context.Context) error
	Enqueue(ctx context.Context, request *domain.UpgradeRequest) error
	ListOperations(ctx context.Context) ([]domain.Operation, error)
	FindOperation(ctx context.Context, operationID string) (*domain.Operation, error)
}

// FileRepository stores state under one directory shared with the worker.
type FileRepository struct {
	stateDir string
}

// NewFileRepository creates a repository rooted at stateDir.
func NewFileRepository(stateDir string) *FileRepository {
	return &FileRepository{stateDir: strings.TrimSpace(stateDir)}
}

func (r *FileRepository) statusPath() string {
	return filepath.Join(r.stateDir, "status.json")
}

func (r *FileRepository) checkPath() string {
	return filepath.Join(r.stateDir, "check-request.json")
}

func (r *FileRepository) queueDir() string {
	return filepath.Join(r.stateDir, "queue")
}

func (r *FileRepository) resultsDir() string {
	return filepath.Join(r.stateDir, "results")
}

// Status returns the worker's latest snapshot. A missing file is reported as
// unavailable so the console can show "unknown" rather than a stale guess.
func (r *FileRepository) Status(ctx context.Context) (*domain.StatusSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	payload, err := os.ReadFile(r.statusPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrStateUnavailable
	}
	if err != nil {
		return nil, fmt.Errorf("read status snapshot: %w", err)
	}
	var snapshot domain.StatusSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return nil, fmt.Errorf("decode status snapshot: %w", err)
	}
	return &snapshot, nil
}

// RequestCheck writes a single check marker that the worker consumes.
func (r *FileRepository) RequestCheck(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	request := domain.CheckRequest{RequestedAt: time.Now().UTC()}
	return writeJSONAtomic(r.checkPath(), request)
}

// Enqueue writes one upgrade command into the worker's queue directory.
func (r *FileRepository) Enqueue(
	ctx context.Context,
	request *domain.UpgradeRequest,
) error {
	if request == nil {
		return errors.New("upgrade request is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(r.queueDir(), 0o750); err != nil {
		return fmt.Errorf("create upgrade queue: %w", err)
	}
	return writeJSONAtomic(
		filepath.Join(r.queueDir(), request.ID+".json"),
		request,
	)
}

// ListOperations returns every known operation, newest first.
func (r *FileRepository) ListOperations(
	ctx context.Context,
) ([]domain.Operation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(r.resultsDir())
	if errors.Is(err, os.ErrNotExist) {
		return []domain.Operation{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list upgrade results: %w", err)
	}
	operations := make([]domain.Operation, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		operation, err := r.readOperation(filepath.Join(r.resultsDir(), entry.Name()))
		if err != nil {
			// A partially written or unreadable file must not break the whole
			// list; operators can still see the remaining history.
			continue
		}
		operations = append(operations, *operation)
	}
	sort.SliceStable(operations, func(left int, right int) bool {
		return operationSortKey(operations[left]) > operationSortKey(operations[right])
	})
	return operations, nil
}

// FindOperation returns one operation by id.
func (r *FileRepository) FindOperation(
	ctx context.Context,
	operationID string,
) (*domain.Operation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" || strings.ContainsAny(operationID, `/\`) {
		return nil, domain.ErrOperationNotFound
	}
	operation, err := r.readOperation(
		filepath.Join(r.resultsDir(), operationID+".json"),
	)
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrOperationNotFound
	}
	if err != nil {
		return nil, err
	}
	return operation, nil
}

func (r *FileRepository) readOperation(path string) (*domain.Operation, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var operation domain.Operation
	if err := json.Unmarshal(payload, &operation); err != nil {
		return nil, fmt.Errorf("decode upgrade result: %w", err)
	}
	return &operation, nil
}

// operationSortKey prefers the most recent timestamp available. Finished
// operations sort by finish time; active operations sort by start time.
func operationSortKey(operation domain.Operation) int64 {
	switch {
	case operation.FinishedAt != nil:
		return operation.FinishedAt.UnixNano()
	case operation.StartedAt != nil:
		return operation.StartedAt.UnixNano()
	default:
		return 0
	}
}

// writeJSONAtomic writes payload to path via a temporary file and rename.
//
// The rename is atomic on the same filesystem, which is why the temporary
// file is created beside the destination instead of in the system temp dir.
func writeJSONAtomic(path string, payload any) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode state file: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".tmp-*.json")
	if err != nil {
		return fmt.Errorf("create state file: %w", err)
	}
	temporaryPath := temporary.Name()
	if _, err := temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("write state file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("close state file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}
