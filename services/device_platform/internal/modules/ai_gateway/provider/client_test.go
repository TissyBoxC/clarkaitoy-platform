package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gatewaydomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
)

func TestCreateAccountSendsConfiguredBalanceAndDecodesProjection(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", request.Method)
		}
		if request.URL.Path != "/internal/sprout/v1/ai-accounts" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"schema_version":"1.0.0",
			"request_id":"request-1",
			"data":{
				"provider_account_id":"parent_abc",
				"user_id":42,
				"status":"active",
				"balance_usd":12.5,
				"concurrency_limit":3,
				"allowed_models":["model-a"]
			},
			"error":null
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, strings.Repeat("x", 32), server.Client())
	if err != nil {
		t.Fatalf("create provider client: %v", err)
	}
	account, err := client.CreateAccount(
		context.Background(),
		gatewaydomain.ProviderAccount{
			ProviderAccountID:    "parent_abc",
			ProviderAccountEmail: "parent@example.com",
			Status:               "active",
			BalanceUSD:           12.5,
			HasBalanceUSD:        true,
			ConcurrencyLimit:     3,
			AllowedModels:        []string{"model-a"},
		},
		"provider-password-long-enough",
	)
	if err != nil {
		t.Fatalf("create provider account: %v", err)
	}
	if received["balance_usd"] != 12.5 {
		t.Fatalf("expected balance_usd 12.5, got %#v", received["balance_usd"])
	}
	if account.BalanceUSD != 12.5 || account.UserID != 42 ||
		account.ProviderAccountID != "parent_abc" {
		t.Fatalf("unexpected provider projection: %#v", account)
	}
}

func TestUpdateAccountOmitsBalanceWhenNotProvided(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"schema_version":"1.0.0",
			"request_id":"request-1",
			"data":{
				"provider_account_id":"parent_abc",
				"user_id":42,
				"status":"active",
				"balance_usd":12.5,
				"concurrency_limit":3,
				"allowed_models":[]
			},
			"error":null
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, strings.Repeat("x", 32), server.Client())
	if err != nil {
		t.Fatalf("create provider client: %v", err)
	}
	if _, err := client.UpdateAccount(
		context.Background(),
		"parent_abc",
		gatewaydomain.ProviderAccount{
			ProviderAccountID: "parent_abc",
			Status:            "active",
			ConcurrencyLimit:  3,
			AllowedModels:     []string{},
		},
		"admin update",
	); err != nil {
		t.Fatalf("update provider account: %v", err)
	}
	if _, exists := received["balance_usd"]; exists {
		t.Fatalf("expected balance_usd to be omitted, got %#v", received["balance_usd"])
	}
}

func TestCreateAccountReturnsProviderRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		_ *http.Request,
	) {
		response.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = response.Write([]byte(`{
			"schema_version":"1.0.0",
			"request_id":"request-1",
			"data":null,
			"error":{"code":"validation_failed","message":"invalid","retryable":false}
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, strings.Repeat("x", 32), server.Client())
	if err != nil {
		t.Fatalf("create provider client: %v", err)
	}
	if _, err := client.CreateAccount(
		context.Background(),
		gatewaydomain.ProviderAccount{},
		"provider-password-long-enough",
	); err == nil || !strings.Contains(err.Error(), gatewaydomain.ErrProviderRejected.Error()) {
		t.Fatalf("expected provider rejection, got %v", err)
	}
}

func TestRotateAPIKeyReturnsReplacementCredential(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", request.Method)
		}
		if request.URL.Path != "/internal/sprout/v1/ai-accounts/parent_abc/api-keys/19/rotate" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"schema_version":"1.0.0",
			"request_id":"request-1",
			"data":{
				"user_id":42,
				"api_key_id":84,
				"name":"sprout-platform",
				"api_key":"replacement-secret",
				"status":"active",
				"quota_usd":0,
				"expires_at":null
			},
			"error":null
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, strings.Repeat("x", 32), server.Client())
	if err != nil {
		t.Fatalf("create provider client: %v", err)
	}
	key, err := client.RotateAPIKey(
		context.Background(),
		"parent_abc",
		19,
		gatewaydomain.ProviderAPIKey{
			Name:     "sprout-platform",
			QuotaUSD: 0,
		},
	)
	if err != nil {
		t.Fatalf("rotate provider key: %v", err)
	}
	if key.ID != 84 || key.UserID != 42 || key.Key != "replacement-secret" {
		t.Fatalf("unexpected replacement credential: %#v", key)
	}
	if received["name"] != "sprout-platform" {
		t.Fatalf("unexpected rotation payload: %#v", received)
	}
}

func TestGetRuntimeConfigDecodesModelsAndDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", request.Method)
		}
		if request.URL.Path != "/internal/sprout/v1/runtime-config" {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"schema_version":"1.0.0",
			"request_id":"request-1",
			"data":{
				"default_balance_usd":3.5,
				"default_concurrency":2,
				"recommended_model":"model-fast",
				"models":[
					{
						"model":"model-fast",
						"status":"operational",
						"primary_latency_ms":120,
						"average_latency_7d_ms":180,
						"recommended_for_new_accounts":true
					}
				]
			},
			"error":null
		}`))
	}))
	defer server.Close()

	client, err := New(server.URL, strings.Repeat("x", 32), server.Client())
	if err != nil {
		t.Fatalf("create provider client: %v", err)
	}
	config, err := client.GetRuntimeConfig(context.Background())
	if err != nil {
		t.Fatalf("get runtime config: %v", err)
	}
	if config.DefaultBalanceUSD != 3.5 ||
		config.DefaultConcurrency != 2 ||
		config.RecommendedModel != "model-fast" {
		t.Fatalf("unexpected runtime defaults: %#v", config)
	}
	if len(config.Models) != 1 ||
		config.Models[0].PrimaryLatencyMs == nil ||
		*config.Models[0].PrimaryLatencyMs != 120 ||
		!config.Models[0].RecommendedForNewAccounts {
		t.Fatalf("unexpected runtime models: %#v", config.Models)
	}
}
