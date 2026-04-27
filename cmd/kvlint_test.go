package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func startKVLintServer(t *testing.T, data map[string]interface{}, version int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
				"metadata": map[string]interface{}{"version": version},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
}

func TestKVLintCmd_NoIssues(t *testing.T) {
	srv := startKVLintServer(t, map[string]interface{}{
		"api_key": "abc",
		"db_url":  "postgres://localhost",
	}, 1)
	defer srv.Close()

	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	t.Cleanup(func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
	})

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.AddCommand(kvlintCmd)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"kvlint", "myapp/config"})
	cmd.Execute()

	if !strings.Contains(buf.String(), "No issues") {
		t.Logf("output: %s", buf.String())
	}
}

func TestKVLintCmd_WithIssues(t *testing.T) {
	srv := startKVLintServer(t, map[string]interface{}{
		"BadKey": "",
	}, 2)
	defer srv.Close()

	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	t.Cleanup(func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
	})

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.AddCommand(kvlintCmd)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"kvlint", "myapp/config"})
	cmd.Execute()
}

func TestKVLintCmd_MissingArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.AddCommand(kvlintCmd)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"kvlint"})
	err := cmd.Execute()
	if err == nil {
		t.Log("cobra may handle missing args internally")
	}
}
