package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

func startKVDeleteServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/secret/delete/myapp" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			return
		}
		if r.URL.Path == "/v1/secret/destroy/myapp" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestKVDeleteCmd_Output(t *testing.T) {
	srv := startKVDeleteServer(t)
	defer srv.Close()
	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")
	buf, cmd := newTestCmd(kvDeleteCmd)
	cmd.SetArgs([]string{"myapp", "1", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(buf.String(), "Deleted versions") {
		t.Errorf("expected deleted message, got: %s", buf.String())
	}
}

func TestKVDeleteCmd_InvalidVersion(t *testing.T) {
	_, cmd := newTestCmd(kvDeleteCmd)
	cmd.SetArgs([]string{"myapp", "notanumber"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func newTestCmd(sub *cobra.Command) (*testBuffer, *cobra.Command) {
	buf := &testBuffer{}
	root := &cobra.Command{Use: "vaultdiff"}
	sub.SetOut(buf)
	root.AddCommand(sub)
	return buf, root
}

type testBuffer struct{ data []byte }

func (b *testBuffer) Write(p []byte) (int, error) { b.data = append(b.data, p...); return len(p), nil }
func (b *testBuffer) String() string              { return string(b.data) }

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}
func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
