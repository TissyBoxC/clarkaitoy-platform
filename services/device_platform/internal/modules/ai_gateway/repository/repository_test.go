package repository

import (
	"reflect"
	"strings"
	"testing"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
)

func TestNonNilStrings(t *testing.T) {
	t.Run("preserves nil as empty array", func(t *testing.T) {
		got := nonNilStrings(nil)
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %v", got)
		}
	})

	t.Run("preserves populated values", func(t *testing.T) {
		want := []string{"model-a", "model-b"}
		got := nonNilStrings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})
}

func TestUpsertProvisioningKeepsCredentialColumnsWritable(t *testing.T) {
	requiredFragments := []string{
		"credential_ciphertext",
		"credential_nonce",
		"provider_api_key_id",
		"ON CONFLICT (parent_account_id) DO UPDATE",
		"RETURNING",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(upsertProvisioningQuery, fragment) {
			t.Fatalf("upsert provisioning query must contain %q", fragment)
		}
	}
}

func TestUpsertProvisioningDoesNotOverwriteCompleteCredential(t *testing.T) {
	if !strings.Contains(
		upsertProvisioningQuery,
		"WHERE ai_accounts.credential_ciphertext = ''::bytea",
	) {
		t.Fatal("upsert must only repair rows with a missing credential")
	}
}

func TestAccountHasCredentialRequiresAllCredentialFields(t *testing.T) {
	valid := &domain.Account{
		APIKeyCiphertext: []byte("ciphertext"),
		APIKeyNonce:      []byte("nonce"),
		ProviderAPIKeyID: 42,
	}
	if !accountHasCredential(valid) {
		t.Fatal("expected a complete credential to be accepted")
	}

	tests := []struct {
		name    string
		account *domain.Account
	}{
		{name: "nil account"},
		{
			name: "missing ciphertext",
			account: &domain.Account{
				APIKeyNonce:      []byte("nonce"),
				ProviderAPIKeyID: 42,
			},
		},
		{
			name: "missing nonce",
			account: &domain.Account{
				APIKeyCiphertext: []byte("ciphertext"),
				ProviderAPIKeyID: 42,
			},
		},
		{
			name: "missing provider key id",
			account: &domain.Account{
				APIKeyCiphertext: []byte("ciphertext"),
				APIKeyNonce:      []byte("nonce"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if accountHasCredential(test.account) {
				t.Fatal("expected an incomplete credential to be rejected")
			}
		})
	}
}
