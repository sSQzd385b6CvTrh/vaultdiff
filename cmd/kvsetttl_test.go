package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func startKVSetTTLServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestKVSetTTLCmd_Set(t *testing.T) {
	svr := startKVSetTTLServer(http.StatusNoContent)
	defer svr.Close()

	t.Setenv("VAULT_ADDR", svr.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	rootCmd.ResetFlags()
	var buf strings.Builder
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{"kv-set-ttl", "myapp/config", "--ttl", "12h"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "TTL set to") {
		t.Errorf("expected TTL set message, got: %q", out)
	}
}

func TestKVSetTTLCmd_Clear(t *testing.T) {
	svr := startKVSetTTLServer(http.StatusNoContent)
	defer svr.Close()

	t.Setenv("VAULT_ADDR", svr.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	rootCmd.ResetFlags()
	var buf strings.Builder
	rootCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{"kv-set-ttl", "myapp/config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "TTL cleared") {
		t.Errorf("expected TTL cleared message, got: %q", out)
	}
}

func TestKVSetTTLCmd_InvalidTTL(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"kv-set-ttl", "myapp/config", "--ttl", "notaduration"})

	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "test-token")

	rootCmd.SetArgs([]string{"kv-set-ttl", "myapp/config", "--ttl", "notaduration"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid TTL")
	}
	if !strings.Contains(err.Error(), "invalid TTL") {
		t.Errorf("expected 'invalid TTL' in error, got: %v", err)
	}
}

func TestKVSetTTLCmd_MissingArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"kv-set-ttl"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}
