package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func kvTagMetadataResponse(tags map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"custom_metadata": tags,
		},
	}
}

func newTagServer(t *testing.T, getHandler, postHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getHandler(w, r)
		case http.MethodPost:
			postHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

func TestGetTags_Success(t *testing.T) {
	expected := map[string]string{"env": "prod", "team": "platform"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kvTagMetadataResponse(expected))
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	tagger := NewKVTagger(NewClient(c), "secret")
	tags, err := tagger.GetTags("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", tags["env"])
	}
	if tags["team"] != "platform" {
		t.Errorf("expected team=platform, got %s", tags["team"])
	}
}

func TestGetTags_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	tagger := NewKVTagger(NewClient(c), "")
	_, err := tagger.GetTags("missing/path")
	if err == nil {
		t.Fatal("expected error for not found path")
	}
}

func TestSetTags_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	tagger := NewKVTagger(NewClient(c), "secret")
	err := tagger.SetTags("myapp/config", map[string]string{"env": "staging"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetTags_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	c, _ := vaultapi.NewClient(cfg)

	tagger := NewKVTagger(NewClient(c), "secret")
	err := tagger.SetTags("myapp/config", map[string]string{"env": "staging"})
	if err == nil {
		t.Fatal("expected error on server error")
	}
}

func TestNewKVTagger_DefaultMount(t *testing.T) {
	cfg := vaultapi.DefaultConfig()
	c, _ := vaultapi.NewClient(cfg)
	tagger := NewKVTagger(NewClient(c), "")
	if tagger.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", tagger.mount)
	}
}
