package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thrasher-corp/vaultdiff/internal/vault"
	"github.com/thrasher-corp/vaultdiff/internal/vault/client"
)

func kvSchemaResponse(data map[string]interface{}) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func newSchemaServer(status int, data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			w.Write(kvSchemaResponse(data))
		}
	}))
}

func TestKVSchema_Valid(t *testing.T) {
	srv := newSchemaServer(http.StatusOK, map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
	})
	defer srv.Close()

	c := &client.Client{Address: srv.URL, Token: "tok", HTTP: srv.Client()}
	v := vault.NewKVSchemaValidator(c, "secret")

	result, err := v.Validate("myapp/db", []vault.SchemaField{
		{Key: "username", Required: true},
		{Key: "password", Required: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got missing=%v", result.Missing)
	}
	if len(result.Extra) != 0 {
		t.Errorf("expected no extra keys, got %v", result.Extra)
	}
}

func TestKVSchema_MissingRequired(t *testing.T) {
	srv := newSchemaServer(http.StatusOK, map[string]interface{}{
		"username": "admin",
	})
	defer srv.Close()

	c := &client.Client{Address: srv.URL, Token: "tok", HTTP: srv.Client()}
	v := vault.NewKVSchemaValidator(c, "secret")

	result, err := v.Validate("myapp/db", []vault.SchemaField{
		{Key: "username", Required: true},
		{Key: "password", Required: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid result")
	}
	if len(result.Missing) != 1 || result.Missing[0] != "password" {
		t.Errorf("expected missing=[password], got %v", result.Missing)
	}
}

func TestKVSchema_ExtraKeys(t *testing.T) {
	srv := newSchemaServer(http.StatusOK, map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
		"debug":    "true",
	})
	defer srv.Close()

	c := &client.Client{Address: srv.URL, Token: "tok", HTTP: srv.Client()}
	v := vault.NewKVSchemaValidator(c, "")

	result, err := v.Validate("myapp/db", []vault.SchemaField{
		{Key: "username", Required: true},
		{Key: "password", Required: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Extra) != 1 || result.Extra[0] != "debug" {
		t.Errorf("expected extra=[debug], got %v", result.Extra)
	}
}

func TestKVSchema_NotFound(t *testing.T) {
	srv := newSchemaServer(http.StatusNotFound, nil)
	defer srv.Close()

	c := &client.Client{Address: srv.URL, Token: "tok", HTTP: srv.Client()}
	v := vault.NewKVSchemaValidator(c, "secret")

	_, err := v.Validate("missing/path", []vault.SchemaField{})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVSchemaValidator_DefaultMount(t *testing.T) {
	c := &client.Client{}
	v := vault.NewKVSchemaValidator(c, "")
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
}
