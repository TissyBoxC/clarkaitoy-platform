package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/TissyBoxC/sprout-platform/packages/go/httpapi"
	authdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	authservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/service"
)

type authHandler struct {
	service *authservice.Service
}

type registerRequest struct {
	Email                  string `json:"email"`
	Password               string `json:"password"`
	DisplayName            string `json:"display_name"`
	Phone                  string `json:"phone"`
	GuardianConsentVersion string `json:"guardian_consent_version"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (handler authHandler) register(response http.ResponseWriter, request *http.Request) {
	var payload registerRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	account, pair, err := handler.service.Register(request.Context(), authdomain.RegisterInput{
		Email:                  payload.Email,
		Password:               payload.Password,
		DisplayName:            payload.DisplayName,
		Phone:                  payload.Phone,
		GuardianConsentVersion: payload.GuardianConsentVersion,
	})
	if err != nil {
		writeAuthServiceError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusCreated, authResponse(account, pair, nil))
}

func (handler authHandler) login(response http.ResponseWriter, request *http.Request) {
	var payload loginRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	account, pair, err := handler.service.Login(request.Context(), authdomain.LoginInput{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		writeAuthServiceError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, authResponse(account, pair, nil))
}

func (handler authHandler) refresh(response http.ResponseWriter, request *http.Request) {
	var payload refreshRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	pair, err := handler.service.Refresh(request.Context(), payload.RefreshToken)
	if err != nil {
		writeAuthServiceError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_in":    int(pair.ExpiresIn.Seconds()),
	})
}

func (handler authHandler) logout(response http.ResponseWriter, request *http.Request) {
	var payload refreshRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeError(response, request, http.StatusBadRequest, "invalid_request", "请检查填写的内容")
		return
	}
	if err := handler.service.Logout(request.Context(), payload.RefreshToken); err != nil {
		writeAuthServiceError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, map[string]any{"logged_out": true})
}

func (handler authHandler) me(response http.ResponseWriter, request *http.Request) {
	accountID, ok := authenticatedAccountID(request)
	if !ok {
		writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
		return
	}
	account, summary, err := handler.service.GetAccount(request.Context(), accountID)
	if err != nil {
		writeAuthServiceError(response, request, err)
		return
	}
	writeSuccess(response, request, http.StatusOK, accountResponse(account, summary))
}

func (handler authHandler) requireAuthentication(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request)
		if !ok {
			writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
			return
		}
		accountID, err := handler.service.VerifyAccessToken(token)
		if err != nil {
			writeError(response, request, http.StatusUnauthorized, "unauthenticated", "请重新登录")
			return
		}
		next(response, request.WithContext(
			context.WithValue(request.Context(), authenticatedAccountKey{}, accountID),
		))
	}
}

func (handler authHandler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return handler.requireAuthentication(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		accountID, _ := authenticatedAccountID(request)
		account, _, err := handler.service.GetAccount(request.Context(), accountID)
		if err != nil || account.Role != authdomain.RoleAdmin {
			writeError(response, request, http.StatusForbidden, "insufficient_permission", "你没有权限访问此页面")
			return
		}
		next(response, request)
	})
}

type authenticatedAccountKey struct{}

func authenticatedAccountID(request *http.Request) (string, bool) {
	accountID, ok := request.Context().Value(authenticatedAccountKey{}).(string)
	return accountID, ok && strings.TrimSpace(accountID) != ""
}

func authResponse(
	account *authdomain.ParentAccount,
	pair *authdomain.TokenPair,
	summary *authdomain.AIAccountSummary,
) map[string]any {
	response := accountResponse(account, summary)
	response["access_token"] = pair.AccessToken
	response["refresh_token"] = pair.RefreshToken
	response["expires_in"] = int(pair.ExpiresIn.Seconds())
	return response
}

func accountResponse(
	account *authdomain.ParentAccount,
	summary *authdomain.AIAccountSummary,
) map[string]any {
	response := map[string]any{
		"account": map[string]any{
			"id":           account.ID,
			"email":        account.Email,
			"display_name": account.DisplayName,
			"role":         account.Role,
			"status":       account.Status,
		},
		"ai_account": nil,
	}
	if summary != nil {
		response["ai_account"] = map[string]any{
			"status":            summary.Status,
			"balance_usd":       summary.BalanceUSD,
			"concurrency_limit": summary.ConcurrencyLimit,
			"allowed_models":    summary.AllowedModels,
			"provider_ready":    summary.ProviderReady,
		}
	}
	return response
}

func writeAuthServiceError(
	response http.ResponseWriter,
	request *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, authdomain.ErrEmailExists):
		writeError(response, request, http.StatusConflict, "email_exists", "这个邮箱已经注册过")
	case errors.Is(err, authdomain.ErrInvalidEmail):
		writeError(response, request, http.StatusUnprocessableEntity, "invalid_email", "请输入有效的邮箱")
	case errors.Is(err, authdomain.ErrWeakPassword):
		writeError(response, request, http.StatusUnprocessableEntity, "weak_password", "密码至少 8 位，并同时包含字母和数字")
	case errors.Is(err, authdomain.ErrInvalidDisplayName):
		writeError(response, request, http.StatusUnprocessableEntity, "invalid_display_name", "请输入家长称呼")
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		writeError(response, request, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码不正确")
	case errors.Is(err, authdomain.ErrAccountDisabled):
		writeError(response, request, http.StatusForbidden, "account_disabled", "这个账号当前无法登录")
	case errors.Is(err, authdomain.ErrSessionNotFound),
		errors.Is(err, authdomain.ErrSessionExpired):
		writeError(response, request, http.StatusUnauthorized, "session_expired", "登录已过期，请重新登录")
	case errors.Is(err, authdomain.ErrAccountNotFound):
		writeError(response, request, http.StatusNotFound, "account_not_found", "没有找到这个账号")
	default:
		writeError(response, request, http.StatusInternalServerError, "service_error", "操作没有完成，请稍后重试")
	}
}

func decodeJSON(request *http.Request, destination any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return nil
}

func bearerToken(request *http.Request) (string, bool) {
	header := strings.TrimSpace(request.Header.Get("Authorization"))
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func writeSuccess(response http.ResponseWriter, request *http.Request, status int, data any) {
	httpapi.WriteSuccess(response, request, status, data)
}

func writeError(response http.ResponseWriter, request *http.Request, status int, code string, message string) {
	httpapi.WriteError(response, request, status, code, message, false)
}
