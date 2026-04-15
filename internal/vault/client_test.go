package vault_test

import (
	"testing"

	"github.com/vaultdiff/vaultdiff/internal/vault"
)

func TestNewClient_DefaultConfig(t *testing.T) {
	client, err := vault.NewClient(vault.Config{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithAddress(t *testing.T) {
	client, err := vault.NewClient(vault.Config{
		Address: "https://vault.example.com:8200",
		Token:   "s.testtoken",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithNamespace(t *testing.T) {
	client, err := vault.NewClient(vault.Config{
		Address:   "https://vault.example.com:8200",
		Token:     "s.testtoken",
		Namespace: "admin/team-a",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSecretVersion_Fields(t *testing.T) {
	sv := vault.SecretVersion{
		Version: 3,
		Data:    map[string]string{"db_password": "secret123"},
		Metadata: map[string]interface{}{
			"created_time": "2024-01-15T10:00:00Z",
		},
	}

	if sv.Version != 3 {
		t.Errorf("expected version 3, got %d", sv.Version)
	}
	if sv.Data["db_password"] != "secret123" {
		t.Errorf("unexpected data value: %s", sv.Data["db_password"])
	}
	if sv.Metadata["created_time"] != "2024-01-15T10:00:00Z" {
		t.Errorf("unexpected metadata value: %v", sv.Metadata["created_time"])
	}
}
