package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func startKVTTLServer(t *testing.T, deleteVersionAfter string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"delete_version_after": deleteVersionAfter,
			},
		})
	}))
	t.Cleanup(ts.Close)
	return ts
}

func TestKVTTLCmd_NoTTL(t *testing.T) {
	ts := startKVTTLServer(t, "0s")

	root := &cobra.Command{Use: "vaultdiff"}
	root.AddCommand(kvttlCmd)

	buf := &bytes.Buffer{}
	kvttlCmd.SetOut(buf)
	kvttlCmd.SetErr(buf)

	kvttlCmd.ResetFlags()
	kvttlCmd.Flags().String("address", ts.URL, "")
	kvttlCmd.Flags().String("token", "test-token", "")
	kvttlCmd.Flags().String("mount", "secret", "")

	root.SetArgs([]string{"kvttl", "myapp/config",
		"--address", ts.URL,
		"--token", "test-token",
		"--mount", "secret",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "none") {
		t.Errorf("expected 'none' in output, got: %s", buf.String())
	}
}

func TestKVTTLCmd_MissingArgs(t *testing.T) {
	root := &cobra.Command{Use: "vaultdiff"}
	root.AddCommand(kvttlCmd)
	root.SetArgs([]string{"kvttl"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}
