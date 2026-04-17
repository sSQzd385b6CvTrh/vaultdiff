package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func startKVPutServer(t *testing.T, version int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"version": version},
		})
	}))
}

func TestKVPutCmd_Output(t *testing.T) {
	ts := startKVPutServer(t, 5)
	defer ts.Close()

	var buf bytes.Buffer
	kvputCmd.SetOut(&buf)
	kvputCmd.SetArgs([]string{"myapp/config", "foo=bar", "--address", ts.URL})
	if err := kvputCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "version 5") {
		t.Errorf("expected version 5 in output, got: %s", out)
	}
	if !strings.Contains(out, "myapp/config") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestKVPutCmd_InvalidPair(t *testing.T) {
	kvputCmd.SetArgs([]string{"myapp/config", "badpair"})
	if err := kvputCmd.Execute(); err == nil {
		t.Fatal("expected error for invalid key=value pair")
	}
}
