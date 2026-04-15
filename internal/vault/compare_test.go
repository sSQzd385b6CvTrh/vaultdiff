package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func compareKVResponse(version int, data map[string]interface{}) []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"data": data,
			"metadata": map[string]interface{}{
				"version": version,
			},
		},
	})
	return body
}

func TestCompare_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			w.Write(compareKVResponse(1, map[string]interface{}{"key": "old"}))
		} else {
			w.Write(compareKVResponse(2, map[string]interface{}{"key": "new"}))
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	fetcher := NewFetcher(client)
	comparer := NewComparer(fetcher)

	pair, err := comparer.Compare("secret", "myapp/config", 1, 2)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}

	if pair.Path != "secret/myapp/config" {
		t.Errorf("Path = %q, want %q", pair.Path, "secret/myapp/config")
	}
	if pair.VersionA != 1 {
		t.Errorf("VersionA = %d, want 1", pair.VersionA)
	}
	if pair.VersionB != 2 {
		t.Errorf("VersionB = %d, want 2", pair.VersionB)
	}
	if fmt.Sprintf("%v", pair.DataA["key"]) != "old" {
		t.Errorf("DataA[key] = %v, want old", pair.DataA["key"])
	}
	if fmt.Sprintf("%v", pair.DataB["key"]) != "new" {
		t.Errorf("DataB[key] = %v, want new", pair.DataB["key"])
	}
}

func TestCompare_FetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	fetcher := NewFetcher(client)
	comparer := NewComparer(fetcher)

	_, err = comparer.Compare("secret", "missing/path", 1, 2)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
