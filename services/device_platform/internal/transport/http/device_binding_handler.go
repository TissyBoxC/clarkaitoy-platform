package http

import (
	"errors"
	"net/http"
	"strings"

	bindingdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/domain"
	bindingservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/service"
)

type deviceBindingHandler struct {
	service *bindingservice.Service
}

type createBindingTokenRequest struct {
	DeviceID string `json:"device_id"`
}

type bindDeviceRequest struct {
	Token           string   `json:"token"`
	DeviceName      string   `json:"device_name"`
	HardwareModel   string   `json:"hardware_model"`
	FirmwareVersion string   `json:"firmware_version"`
	Capabilities    []string `json:"capabilities"`
}

func (handler deviceBindingHandler) createToken(
	response http.ResponseWriter,
	request *http.Request,
) {
	var payload createBindingTokenRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查设备信息")
		return
	}
	token, details, err := handler.service.CreateToken(request.Context(), payload.DeviceID)
	if err != nil {
		writeDeviceBindingError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusCreated, map[string]any{
		"token":      token,
		"device_id":  details.DeviceID,
		"expires_at": details.ExpiresAt,
	})
}

func (handler deviceBindingHandler) bind(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	var payload bindDeviceRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查设备信息")
		return
	}
	binding, err := handler.service.Bind(
		request.Context(),
		accountID,
		payload.Token,
		payload.DeviceName,
		payload.HardwareModel,
		payload.FirmwareVersion,
		payload.Capabilities,
	)
	if err != nil {
		writeDeviceBindingError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusCreated, bindingResponse(binding))
}

func (handler deviceBindingHandler) list(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	bindings, err := handler.service.List(request.Context(), accountID)
	if err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取设备")
		return
	}
	result := make([]map[string]any, 0, len(bindings))
	for index := range bindings {
		result = append(result, bindingResponse(&bindings[index]))
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{"devices": result})
}

func (handler deviceBindingHandler) remove(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	deviceID := strings.TrimSpace(request.PathValue("device_id"))
	if err := handler.service.Delete(request.Context(), accountID, deviceID); err != nil {
		writeDeviceBindingError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{"removed": true})
}

func bindingResponse(binding *bindingdomain.Binding) map[string]any {
	return map[string]any{
		"device_id":        binding.DeviceID,
		"device_name":      binding.DeviceName,
		"hardware_model":   binding.HardwareModel,
		"firmware_version": binding.FirmwareVersion,
		"capabilities":     binding.Capabilities,
		"bound_at":         binding.BoundAt,
		"updated_at":       binding.UpdatedAt,
	}
}

func writeDeviceBindingError(
	response http.ResponseWriter,
	request *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, bindingdomain.ErrInvalidDeviceID):
		writeError(response, request, http.StatusUnprocessableEntity, "invalid_device", "设备信息不正确")
	case errors.Is(err, bindingdomain.ErrTokenNotFound):
		writeError(response, request, http.StatusNotFound, "token_not_found", "绑定码无效，请在设备上重新生成")
	case errors.Is(err, bindingdomain.ErrTokenExpired):
		writeError(response, request, http.StatusGone, "token_expired", "绑定码已过期，请在设备上重新生成")
	case errors.Is(err, bindingdomain.ErrTokenConsumed):
		writeError(response, request, http.StatusConflict, "token_used", "这个绑定码已经使用，请在设备上重新生成")
	case errors.Is(err, bindingdomain.ErrDeviceNotFound):
		writeError(response, request, http.StatusNotFound, "device_not_found", "没有找到这台设备")
	default:
		writeError(response, request, http.StatusInternalServerError, "service_error", "操作没有完成，请稍后重试")
	}
}
