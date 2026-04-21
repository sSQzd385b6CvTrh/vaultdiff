package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func startRestoreVersionServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

func TestKVRestoreVersionCmd_Output(t *testing.T) {
	srv := startRestoreVersionServer(t, http.StatusNoContent)
	defer srv.Close()

	os.Setenv("VAULT_ADDR", srv.URL)
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_ADDR")
	defer os.Unsetenv("VAULT_TOKEN")

	root := &cobra.Command{Use: "vaultdiff"}
	buf := &bytes.Buffer{}
	root.SetOut(buf)

	for _, sub := range rootCmd.Commands() {
		if sub.Use == "kv-restore-version <path> <version>[,version...]" {
			root.AddCommand(sub)
		}
	}

	root.SetArgs([]string{"kv-restore-version", "myapp/config", "2,3"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKVRestoreVersionCmd_InvalidVersion(t *testing.T) {
	err := parseVersionInts("abc")
	if _, ok := err.(error); ok {
		// expected
	}
	_, err2 := parseVersionInts("abc")
	if err2 == nil {
		t.Fatal("expected error for non-numeric version")
	}
}

func TestParseVersionInts_Multiple(t *testing.T) {
	versions, err := parseVersionInts("1,2,3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}
}

func TestParseVersionInts_Empty(t *testing.T) {
	_, err := parseVersionInts("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}
