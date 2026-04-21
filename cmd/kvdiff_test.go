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

func startKVDiffServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var data map[string]interface{}
		if strings.Contains(r.URL.Path, "src") {
			data = map[string]interface{}{"password": "old", "host": "localhost"}
		} else {
			data = map[string]interface{}{"password": "new", "host": "localhost"}
		}
		body, _ := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
}

func TestKVDiffCmd_Output(t *testing.T) {
	server := startKVDiffServer(t)
	defer server.Close()

	os.Setenv("VAULT_ADDR", server.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	buf := &bytes.Buffer{}
	kvdiffCmd.SetOut(buf)
	kvdiffCmd.SetArgs([]string{"src/app", "dst/app"})

	err := kvdiffCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "password") {
		t.Errorf("expected 'password' in output, got: %s", out)
	}
}

func TestKVDiffCmd_MissingArgs(t *testing.T) {
	kvdiffCmd.SetArgs([]string{"only-one-arg"})
	err := kvdiffCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing second arg")
	}
}
