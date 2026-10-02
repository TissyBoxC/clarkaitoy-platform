package http

import (
	"errors"
	"net/http"
	"strings"

	gatewaydomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
	gatewayservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/service"
)

type adminHandler struct {
	service *gatewayservice.Service
}

type updateAIAccountRequest struct {
	Status           string   `json:"status"`
	BalanceUSD       float64  `json:"balance_usd"`
	ConcurrencyLimit int      `json:"concurrency_limit"`
	AvailableModels  []string `json:"available_models"`
	Reason           string   `json:"reason"`
}

type updateAIModelsRequest struct {
	SelectedModels []string `json:"selected_models"`
}

func (handler adminHandler) listAIAccounts(
	response http.ResponseWriter,
	request *http.Request,
) {
	accounts, err := handler.service.ListForAdmin(request.Context())
	if err != nil {
		writeError(response, request, http.StatusInternalServerError, "service_error", "暂时无法读取家长账号")
		return
	}
	result := make([]map[string]any, 0, len(accounts))
	for index := range accounts {
		result = append(result, aiAccountAdminResponse(&accounts[index]))
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{"accounts": result})
}

func (handler adminHandler) updateAIAccount(
	response http.ResponseWriter,
	request *http.Request,
) {
	providerAccountID := strings.TrimSpace(request.PathValue("provider_account_id"))
	if providerAccountID == "" {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	var payload updateAIAccountRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	if payload.Status != "active" && payload.Status != "suspended" {
		writeError(response, request, http.StatusUnprocessableEntity, "invalid_status", "请选择有效的账号状态")
		return
	}
	if payload.BalanceUSD < 0 || payload.ConcurrencyLimit < 1 {
		writeError(response, request, http.StatusUnprocessableEntity, "invalid_limits", "余额和并发额度必须大于等于 0")
		return
	}
	account, err := handler.service.UpdateForAdmin(
		request.Context(),
		providerAccountID,
		payload.Status,
		payload.BalanceUSD,
		payload.ConcurrencyLimit,
		normalizeModels(payload.AvailableModels),
		payload.Reason,
	)
	if err != nil {
		if errors.Is(err, gatewaydomain.ErrAccountNotFound) {
			writeError(response, request, http.StatusNotFound, "not_found", "没有找到这个账号")
			return
		}
		writeError(response, request, http.StatusBadGateway, "provider_error", "暂时无法更新 AI 服务，请稍后重试")
		return
	}
	writeSuccess(response, request, http.StatusOK, aiAccountAdminResponse(account))
}

// updateParentAIModels lets a guardian choose from provider-approved models.
// The response is a parent-safe summary and never contains a provider key.
func (handler adminHandler) updateParentAIModels(
	response http.ResponseWriter,
	request *http.Request,
) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	var payload updateAIModelsRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查选择的模型")
		return
	}
	summary, err := handler.service.UpdateModelsForParent(
		request.Context(),
		accountID,
		payload.SelectedModels,
	)
	if err != nil {
		switch {
		case errors.Is(err, gatewaydomain.ErrModelNotAllowed):
			writeError(response, request, http.StatusUnprocessableEntity, "model_not_available", "请选择当前可用的模型")
		case errors.Is(err, gatewaydomain.ErrAccountNotFound):
			writeError(response, request, http.StatusNotFound, "ai_account_not_found", "还没有可用的 AI 服务")
		case errors.Is(err, gatewaydomain.ErrProviderUnavailable),
			errors.Is(err, gatewaydomain.ErrProviderRejected):
			writeError(response, request, http.StatusBadGateway, "ai_service_unavailable", "AI 服务暂时不可用，请稍后重试")
		default:
			writeError(response, request, http.StatusInternalServerError, "service_error", "操作没有完成，请稍后重试")
		}
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{
		"ai_account": map[string]any{
			"status":            summary.Status,
			"balance_usd":       summary.BalanceUSD,
			"concurrency_limit": summary.ConcurrencyLimit,
			"available_models":  summary.AvailableModels,
			"selected_models":   summary.SelectedModels,
			"allowed_models":    summary.AllowedModels,
			"provider_ready":    summary.ProviderReady,
		},
	})
}

func aiAccountAdminResponse(account *gatewaydomain.Account) map[string]any {
	return map[string]any{
		"parent_account_id":   account.ParentAccountID,
		"parent_email":        account.ParentEmail,
		"parent_display_name": account.ParentDisplayName,
		"provider_account_id": account.ProviderAccountID,
		"status":              account.Status,
		"balance_usd":         account.BalanceUSD,
		"concurrency_limit":   account.ConcurrencyLimit,
		"available_models":    account.AvailableModels,
		"selected_models":     account.SelectedModels,
		"allowed_models":      account.AllowedModels,
		"credential_ready":    len(account.APIKeyCiphertext) > 0,
		"updated_at":          account.UpdatedAt,
	}
}

func normalizeModels(models []string) []string {
	normalized := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" || len(model) > 128 {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		normalized = append(normalized, model)
	}
	return normalized
}
