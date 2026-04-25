package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func annotateMetadataResponse(meta map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"custom_metadata": meta,
		},
	}
}

func newAnnotateServer(t *testing.T, getResp map[string]interface{}, getStatus int, postStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(getStatus)
			if getResp != nil {
				json.NewEncoder(w).Encode(getResp)
			}
			return
		}
		w.WriteHeader(postStatus)
	}))
}

func TestGetAnnotations_Success(t *testing.T) {
	resp := annotateMetadataResponse(map[string]string{"owner": "team-a", "env": "prod"})
	srv := newAnnotateServer(t, resp, http.StatusOK, http.StatusNoContent)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	a := NewKVAnnotator(client, "secret")
	ann, err := a.GetAnnotations("myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ann.Annotations["owner"] != "team-a" {
		t.Errorf("expected owner=team-a, got %s", ann.Annotations["owner"])
	}
}

func TestGetAnnotations_NotFound(t *testing.T) {
	srv := newAnnotateServer(t, nil, http.StatusNotFound, http.StatusNoContent)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	a := NewKVAnnotator(client, "secret")
	_, err := a.GetAnnotations("missing/path")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetAnnotations_Success(t *testing.T) {
	srv := newAnnotateServer(t, nil, http.StatusOK, http.StatusNoContent)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	a := NewKVAnnotator(client, "secret")
	err := a.SetAnnotations("myapp/config", map[string]string{"owner": "team-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAnnotations_ServerError(t *testing.T) {
	srv := newAnnotateServer(t, nil, http.StatusOK, http.StatusInternalServerError)
	defer srv.Close()

	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := api.NewClient(cfg)

	a := NewKVAnnotator(client, "secret")
	err := a.SetAnnotations("myapp/config", map[string]string{"owner": "team-b"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewKVAnnotator_DefaultMount(t *testing.T) {
	cfg := api.DefaultConfig()
	client, _ := api.NewClient(cfg)
	a := NewKVAnnotator(client, "")
	if a.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", a.mount)
	}
}
