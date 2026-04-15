package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func envKVResponse(data map[string]interface{}) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{"version": 1},
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func TestEnvComparer_Compare_Success(t *testing.T) {
	srcData := map[string]interface{}{"key": "alpha"}
	dstData := map[string]interface{}{"key": "beta"}

	srcServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(envKVResponse(srcData))
	}))
	defer srcServer.Close()

	dstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(envKVResponse(dstData))
	}))
	defer dstServer.Close()

	ec, err := NewEnvComparer(map[string]EnvConfig{
		"staging": {Address: srcServer.URL, Token: "tok"},
		"prod":    {Address: dstServer.URL, Token: "tok"},
	})
	if err != nil {
		t.Fatalf("NewEnvComparer: %v", err)
	}

	src, dst, err := ec.Compare(context.Background(), "staging", "prod", "secret/app", 1, 1)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if src["key"] != "alpha" {
		t.Errorf("expected src key=alpha, got %v", src["key"])
	}
	if dst["key"] != "beta" {
		t.Errorf("expected dst key=beta, got %v", dst["key"])
	}
}

func TestEnvComparer_UnknownEnv(t *testing.T) {
	ec := &EnvComparer{envs: map[string]*Fetcher{}}
	_, _, err := ec.Compare(context.Background(), "missing", "prod", "secret/app", 1, 1)
	if err == nil {
		t.Fatal("expected error for unknown src env")
	}
}

func TestNewEnvComparer_InvalidAddress(t *testing.T) {
	_, err := NewEnvComparer(map[string]EnvConfig{
		"bad": {Address: "://invalid", Token: "tok"},
	})
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
}
