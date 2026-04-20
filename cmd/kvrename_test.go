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

func startKVRenameServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "data/old-key"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"foo": "bar"},
				},
			})
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "data/new-key"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "metadata/old-key"):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestKVRenameCmd_Output(t *testing.T) {
	srv := startKVRenameServer(t)
	defer srv.Close()

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "root"}
	kvRenameCmd.ResetFlags()
	kvRenameCmd.Flags().String("address", "", "")
	kvRenameCmd.Flags().String("token", "", "")
	kvRenameCmd.Flags().String("mount", "secret", "")
	cmd.AddCommand(kvRenameCmd)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"kv-rename", "old-key", "new-key",
		"--address", srv.URL, "--token", "test-token"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVRenameCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	cmd.AddCommand(kvRenameCmd)
	cmd.SetArgs([]string{"kv-rename", "only-one-arg"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing dst-path argument")
	}
}
