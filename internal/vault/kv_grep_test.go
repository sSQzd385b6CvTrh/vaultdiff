package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvGrepResponse(data map[string]interface{}) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func newGrepServer(status int, data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			w.Write(kvGrepResponse(data))
		}
	}))
}

func TestKVGrep_MatchValue(t *testing.T) {
	ts := newGrepServer(http.StatusOK, map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
	})
	defer ts.Close()

	client := &Client{Address: ts.URL, Token: "tok", HTTP: ts.Client()}
	g := NewKVGrepper(client, "secret")
	res, err := g.Grep(context.Background(), "myapp/config", "s3cr3t", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected result, got nil")
	}
	if _, ok := res.Matches["password"]; !ok {
		t.Error("expected 'password' in matches")
	}
	if _, ok := res.Matches["username"]; ok {
		t.Error("did not expect 'username' in matches")
	}
}

func TestKVGrep_MatchKey(t *testing.T) {
	ts := newGrepServer(http.StatusOK, map[string]interface{}{
		"db_password": "abc",
		"api_key":     "xyz",
	})
	defer ts.Close()

	client := &Client{Address: ts.URL, Token: "tok", HTTP: ts.Client()}
	g := NewKVGrepper(client, "secret")
	res, err := g.Grep(context.Background(), "myapp/db", "password", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || len(res.Matches) != 1 {
		t.Fatalf("expected 1 match, got %v", res)
	}
}

func TestKVGrep_NotFound(t *testing.T) {
	ts := newGrepServer(http.StatusNotFound, nil)
	defer ts.Close()

	client := &Client{Address: ts.URL, Token: "tok", HTTP: ts.Client()}
	g := NewKVGrepper(client, "secret")
	_, err := g.Grep(context.Background(), "missing/path", "x", false)
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVGrepper_DefaultMount(t *testing.T) {
	client := &Client{}
	g := NewKVGrepper(client, "")
	if g.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", g.mount)
	}
}
