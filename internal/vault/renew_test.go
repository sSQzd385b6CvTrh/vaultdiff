package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func renewResponse(token string, duration int, renewable bool) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"auth": map[string]interface{}{
			"client_token":   token,
			"lease_duration": duration,
			"renewable":      renewable,
		},
	})
	return body
}

func TestRenew_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/renew-self" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(renewResponse("newtoken123", 3600, true))
	}))
	defer ts.Close()

	r := NewTokenRenewer(ts.URL, "mytoken")
	result, err := r.Renew()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ClientToken != "newtoken123" {
		t.Errorf("expected newtoken123, got %s", result.ClientToken)
	}
	if result.LeaseDuration.Seconds() != 3600 {
		t.Errorf("expected 3600s, got %v", result.LeaseDuration)
	}
	if !result.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestRenew_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	r := NewTokenRenewer(ts.URL, "badtoken")
	_, err := r.Renew()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenew_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := NewTokenRenewer(ts.URL, "token")
	_, err := r.Renew()
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
