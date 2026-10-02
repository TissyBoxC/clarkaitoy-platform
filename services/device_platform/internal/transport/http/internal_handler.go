package http

import (
	"errors"
	"net/http"
	"strings"

	gatewaydomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
	gatewayservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/service"
	bindingdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/domain"
	bindingservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/service"
)

// internalHandler exposes service-to-service endpoints for the AI relay.
// Every route must be mounted behind the internal service token.
type internalHandler struct {
	bindingService *bindingservice.Service
	aiService      *gatewayservice.Service
}

// aiCredential resolves the AI execution credential for one bound device.
//
// The relay asks by device id because that is the only identity it already
// holds. The platform resolves the owning guardian and returns the decrypted
// provider key, which must never reach a guardian, admin, or firmware client.
func (handler internalHandler) aiCredential(
	response http.ResponseWriter,
	request *http.Request,
) {
	deviceID := strings.TrimSpace(request.PathValue("device_id"))
	binding, err := handler.bindingService.GetByDeviceID(request.Context(), deviceID)
	if err != nil {
		writeInternalCredentialError(response, request, err)
		return
	}
	credential, err := handler.aiService.CredentialForDevice(
		request.Context(),
		binding.ParentAccountID,
	)
	if err != nil {
		writeInternalCredentialError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{
		"device_id":           binding.DeviceID,
		"provider_account_id": credential.ProviderAccountID,
		"api_key":             credential.APIKey,
	})
}

func writeInternalCredentialError(
	response http.ResponseWriter,
	request *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, bindingdomain.ErrInvalidDeviceID),
		errors.Is(err, bindingdomain.ErrDeviceNotFound):
		writeError(response, request, http.StatusNotFound, "device_not_found", "没有找到这台设备")
	case errors.Is(err, gatewaydomain.ErrAccountNotFound),
		errors.Is(err, gatewaydomain.ErrCredentialInvalid):
		writeError(response, request, http.StatusNotFound, "credential_not_found", "这台设备还没有可用的 AI 服务")
	case errors.Is(err, gatewaydomain.ErrProviderUnavailable):
		writeError(response, request, http.StatusBadGateway, "ai_service_unavailable", "AI 服务暂时不可用，请稍后重试")
	default:
		writeError(response, request, http.StatusInternalServerError, "service_error", "操作没有完成，请稍后重试")
	}
}
