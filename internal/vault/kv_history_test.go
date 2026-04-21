package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func kvHistoryResponse(versions map[string]interface{}) []byte {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"versions": versions,
		},
	}
	b, _ := json.Marshal(body)
	return b
}

func TestGetHistory_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	versions := map[string]interface{}{
		"1": map[string]interface{}{"created_time": now.Format(time.RFC3339), "deletion_time": "", "destroyed": false},
		"2": map[string]interface{}{"created_time": now.Add(time.Minute).Format(time.RFC3339), "deletion_time": "", "destroyed": false},
		"3": map[string]interface{}{"created_time": now.Add(2 * time.Minute).Format(time.RFC3339), "deletion_time": "", "destroyed": true},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/metadata/myapp/db" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(kvHistoryResponse(versions))
	}))
	defer ts.Close()

	client := &Client{Address: ts.URL, Token: "test", HTTP: ts.Client()}
	reader := NewKVHistoryReader(client, "secret")

	history, err := reader.GetHistory(context.Background(), "myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(history))
	}
	if history[0].Version != 1 {
		t.Errorf("expected first version to be 1, got %d", history[0].Version)
	}
	if !history[2].Destroyed {
		t.Errorf("expected version 3 to be destroyed")
	}
}

func TestGetHistory_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := &Client{Address: ts.URL, Token: "test", HTTP: ts.Client()}
	reader := NewKVHistoryReader(client, "secret")

	_, err := reader.GetHistory(context.Background(), "missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestNewKVHistoryReader_DefaultMount(t *testing.T) {
	client := &Client{}
	reader := NewKVHistoryReader(client, "")
	if reader.mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", reader.mount)
	}
}
