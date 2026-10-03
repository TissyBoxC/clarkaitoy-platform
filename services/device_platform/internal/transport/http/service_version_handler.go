package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/service_version/domain"
)

// versionSnapshotResponse is the console-facing service inventory.
type versionSnapshotResponse struct {
	Services    []domain.Service `json:"services"`
	CheckedAt   *time.Time       `json:"checked_at"`
	AllUpToDate bool             `json:"all_up_to_date"`
}

func (handler adminHandler) listServiceVersions(
	response http.ResponseWriter,
	request *http.Request,
) {
	services, checkedAt, err := handler.serviceVersionService.Snapshot(request.Context())
	if err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取服务版本")
		return
	}
	writeSuccess(response, request, http.StatusOK, versionSnapshotResponse{
		Services:    services,
		CheckedAt:   timePointer(checkedAt),
		AllUpToDate: handler.serviceVersionService.AllCurrent(services),
	})
}

func (handler adminHandler) checkServiceVersions(
	response http.ResponseWriter,
	request *http.Request,
) {
	if err := handler.serviceVersionService.RequestCheck(request.Context()); err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "检查更新没有开始，请稍后重试")
		return
	}
	// The worker refreshes asynchronously; return the current known snapshot
	// immediately so the console can render without waiting on the network.
	services, checkedAt, err := handler.serviceVersionService.Snapshot(request.Context())
	if err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取服务版本")
		return
	}
	writeSuccess(response, request, http.StatusOK, versionSnapshotResponse{
		Services:    services,
		CheckedAt:   timePointer(checkedAt),
		AllUpToDate: handler.serviceVersionService.AllCurrent(services),
	})
}

func (handler adminHandler) upgradeService(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	serviceID := strings.TrimSpace(request.PathValue("service"))
	plan, err := handler.serviceVersionService.PlanUpgrade(request.Context(), serviceID)
	if err != nil {
		writeServiceVersionError(response, request, err)
		return
	}
	operation, err := handler.serviceVersionService.Enqueue(
		request.Context(),
		plan,
		accountID,
	)
	if err != nil {
		writeServiceVersionError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusAccepted, map[string]any{
		"operation": operation,
	})
}

func (handler adminHandler) upgradeAllServices(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	plans, err := handler.serviceVersionService.PlanUpgradeAll(request.Context())
	if err != nil {
		writeServiceVersionError(response, request, err)
		return
	}
	operations, err := handler.serviceVersionService.EnqueueAll(
		request.Context(),
		plans,
		accountID,
	)
	if err != nil {
		writeServiceVersionError(response, request, err)
		return
	}
	// The frontend tracks a single batch through the first operation id and
	// reads the rest from the operation list.
	var first any
	if len(operations) > 0 {
		first = operations[0]
	}
	writeSuccess(response, request, http.StatusAccepted, map[string]any{
		"operation":  first,
		"operations": operations,
	})
}

func (handler adminHandler) listServiceVersionOperations(
	response http.ResponseWriter,
	request *http.Request,
) {
	operations, err := handler.serviceVersionService.ListOperations(request.Context())
	if err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取升级记录")
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{
		"operations": operations,
	})
}

func (handler adminHandler) getServiceVersionOperation(
	response http.ResponseWriter,
	request *http.Request,
) {
	operationID := strings.TrimSpace(request.PathValue("operation_id"))
	operation, err := handler.serviceVersionService.FindOperation(
		request.Context(),
		operationID,
	)
	if err != nil {
		if errors.Is(err, domain.ErrOperationNotFound) {
			writeError(response, request, http.StatusNotFound, "operation_not_found", "没有找到这个升级任务")
			return
		}
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取升级进度")
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{
		"operation": operation,
	})
}

func writeServiceVersionError(
	response http.ResponseWriter,
	request *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, domain.ErrServiceNotFound):
		writeError(response, request, http.StatusNotFound, "service_not_found", "没有找到这个服务")
	case errors.Is(err, domain.ErrServiceNotUpgradable):
		writeError(response, request, http.StatusUnprocessableEntity, "service_not_upgradable", "这个服务当前不需要升级")
	case errors.Is(err, domain.ErrNoUpgradeAvailable):
		writeError(response, request, http.StatusConflict, "already_current", "所有服务都已是最新版本")
	case errors.Is(err, domain.ErrUpgradeInProgress):
		writeError(response, request, http.StatusConflict, "upgrade_in_progress", "已有升级任务正在进行，请等待完成")
	default:
		writeError(response, request, http.StatusInternalServerError, "service_error", "升级任务没有开始，请稍后重试")
	}
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
