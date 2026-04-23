package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func startKVAuditTrailServer(t *testing.T) *httptest.Server {
	t.Helper()
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"versions": map[string]interface{}{
				"1": map[string]interface{}{
					"created_time":  "2024-01-15T10:00:00Z",
					"deletion_time": "",
					"destroyed":     false,
				},
				"2": map[string]interface{}{
					"created_time":  "2024-02-20T12:30:00Z",
					"deletion_time": "",
					"destroyed":     true,
				},
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(body)
	}))
}

func TestKVAuditTrailCmd_Output(t *testing.T) {
	srv := startKVAuditTrailServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	buf := &bytes.Buffer{}
	kvAuditTrailCmd.SetOut(buf)
	kvAuditTrailCmd.SetErr(buf)

	kvAuditTrailCmd.SetArgs([]string{"myapp/config"})
	if err := kvAuditTrailCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "VERSION") {
		t.Error("expected header row in output")
	}
	if !strings.Contains(out, "2024-01-15") {
		t.Error("expected version 1 created time in output")
	}
	if !strings.Contains(out, "true") {
		t.Error("expected destroyed=true for version 2")
	}
}

func TestKVAuditTrailCmd_MissingArgs(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	kvAuditTrailCmd.SetArgs([]string{})
	err := kvAuditTrailCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}
