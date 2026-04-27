package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startKVBulkMoveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
			})
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestKVBulkMoveCmd_Output(t *testing.T) {
	srv := startKVBulkMoveServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"kvbulkmove", "src/a=dst/a", "src/b=dst/b"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if out == "" {
		t.Error("expected output, got empty string")
	}
}

func TestKVBulkMoveCmd_InvalidPair(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "test-token")

	rootCmd.SetArgs([]string{"kvbulkmove", "invalidsyntax"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("expected error for invalid pair syntax")
	}
}

func TestParseBulkMovePairs_Valid(t *testing.T) {
	pairs, err := parseBulkMovePairs([]string{"a/b=c/d", "x=y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Source != "a/b" || pairs[0].Dest != "c/d" {
		t.Errorf("unexpected first pair: %+v", pairs[0])
	}
}

func TestParseBulkMovePairs_Invalid(t *testing.T) {
	_, err := parseBulkMovePairs([]string{"nodest"})
	if err == nil {
		t.Error("expected error for missing dest")
	}
}
