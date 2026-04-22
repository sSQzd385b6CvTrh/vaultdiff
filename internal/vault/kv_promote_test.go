package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func promoteKVResponse(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 1},
		},
	}
}

func newPromoteServer(t *testing.T, srcData map[string]interface{}, wantDst string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(promoteKVResponse(srcData))
		case r.Method == http.MethodPost:
			if r.URL.Path != "/v1/secret/data/"+wantDst {
				http.Error(w, "wrong dst", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 2}})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

func TestKVPromote_Success(t *testing.T) {
	srcData := map[string]interface{}{"api_key": "abc123", "db_pass": "secret"}
	srv := newPromoteServer(t, srcData, "prod/myapp")
	defer srv.Close()

	c := clientForURL(srv.URL)
	p := NewKVPromoter(c, "secret", false)

	res, err := p.Promote(context.Background(), "staging/myapp", "prod/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SourcePath != "staging/myapp" {
		t.Errorf("expected source staging/myapp, got %s", res.SourcePath)
	}
	if res.DestPath != "prod/myapp" {
		t.Errorf("expected dest prod/myapp, got %s", res.DestPath)
	}
	if res.Data["api_key"] != "abc123" {
		t.Errorf("expected api_key abc123, got %v", res.Data["api_key"])
	}
}

func TestKVPromote_SourceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := clientForURL(srv.URL)
	p := NewKVPromoter(c, "secret", false)

	_, err := p.Promote(context.Background(), "staging/missing", "prod/missing")
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestNewKVPromoter_DefaultMount(t *testing.T) {
	c := &Client{}
	p := NewKVPromoter(c, "", false)
	if p.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", p.mount)
	}
}
