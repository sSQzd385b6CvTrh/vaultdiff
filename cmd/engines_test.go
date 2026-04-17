package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func startEnginesServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/mounts" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"cubbyhole/": map[string]interface{}{
					"type":        "cubbyhole",
					"description": "per-token private secret storage",
					"accessor":    "cubbyhole_xyz",
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestEnginesCmd_Output(t *testing.T) {
	ts := startEnginesServer(t)
	defer ts.Close()

	t.Setenv("VAULT_ADDR", ts.URL)
	t.Setenv("VAULT_TOKEN", "test-token")
	t.Setenv("VAULT_NAMESPACE", "")

	buf := &bytes.Buffer{}
	enginesCmd.SetOut(buf)
	enginesCmd.SetErr(buf)

	err := enginesCmd.RunE(enginesCmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "cubbyhole/") {
		t.Errorf("expected cubbyhole/ in output, got: %s", out)
	}
	if !strings.Contains(out, "cubbyhole") {
		t.Errorf("expected type cubbyhole in output, got: %s", out)
	}
}
