package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func startHealthServer(t *testing.T, initialized, sealed, standby bool, version string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"initialized":  initialized,
			"sealed":       sealed,
			"standby":      standby,
			"version":      version,
			"cluster_name": "test-cluster",
			"cluster_id":   "xyz-789",
		})
	}))
}

func TestHealthCmd_Output(t *testing.T) {
	srv := startHealthServer(t, true, false, false, "1.15.0")
	defer srv.Close()

	buf := &bytes.Buffer{}
	healthCmd.SetOut(buf)
	healthCmd.SetArgs([]string{"--address", srv.URL})

	if err := healthCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Initialized", "true", "1.15.0", "test-cluster"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestHealthCmd_SealedWarning(t *testing.T) {
	srv := startHealthServer(t, true, true, false, "1.15.0")
	defer srv.Close()

	buf := &bytes.Buffer{}
	healthCmd.SetOut(buf)
	healthCmd.SetArgs([]string{"--address", srv.URL})

	if err := healthCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "WARNING") {
		t.Error("expected sealed warning in output")
	}
}
