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

func startSnapshotDiffServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		switch r.URL.Path {
		case "/v1/secret/data/src":
			body = map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"alpha": "1", "beta": "old"},
				},
			}
		case "/v1/secret/data/dst":
			body = map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"alpha": "1", "beta": "new", "gamma": "added"},
				},
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
}

func TestKVSnapshotDiffCmd_Output(t *testing.T) {
	srv := startSnapshotDiffServer(t)
	defer srv.Close()

	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	kvSnapshotDiffCmd.SetOut(buf)
	kvSnapshotDiffCmd.SetArgs([]string{"src", "dst", "--mount", "secret"})

	if err := kvSnapshotDiffCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "added") {
		t.Errorf("expected 'added' section in output")
	}
	if !strings.Contains(out, "gamma") {
		t.Errorf("expected 'gamma' in added output")
	}
	if !strings.Contains(out, "modified") {
		t.Errorf("expected 'modified' section in output")
	}
	if !strings.Contains(out, "beta") {
		t.Errorf("expected 'beta' in modified output")
	}
}

func TestKVSnapshotDiffCmd_MissingArgs(t *testing.T) {
	kvSnapshotDiffCmd.SetArgs([]string{})
	err := kvSnapshotDiffCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}
