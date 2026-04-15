package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func policyListResponse(keys []string) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{"keys": keys},
	})
	return body
}

func policyGetResponse(name, rules string) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{"name": name, "policy": rules},
	})
	return body
}

func TestListPolicies_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(policyListResponse([]string{"default", "root", "dev-readonly"}))
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lister := NewPolicyLister(c)
	names, err := lister.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListPolicies: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 policies, got %d", len(names))
	}
}

func TestListPolicies_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{}}`))
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lister := NewPolicyLister(c)
	names, err := lister.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 policies, got %d", len(names))
	}
}

func TestGetPolicy_Success(t *testing.T) {
	expectedRules := `path "secret/*" { capabilities = ["read"] }`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(policyGetResponse("dev-readonly", expectedRules))
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lister := NewPolicyLister(c)
	entry, err := lister.GetPolicy(context.Background(), "dev-readonly")
	if err != nil {
		t.Fatalf("GetPolicy: %v", err)
	}
	if entry.Rules != expectedRules {
		t.Errorf("expected rules %q, got %q", expectedRules, entry.Rules)
	}
}
