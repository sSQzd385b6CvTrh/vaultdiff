package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kvEnvDiffResponse(data map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}
}

func newEnvDiffServer(leftData, rightData map[string]interface{}) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/secret/data/myapp", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if host == "left" {
			kvEnvDiffResponse(leftData)(w, r)
		} else {
			kvEnvDiffResponse(rightData)(w, r)
		}
	})
	return httptest.NewServer(mux)
}

func makeEnvGetter(t *testing.T, addr string) *KVGetter {
	t.Helper()
	c, err := NewClient(addr, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return NewKVGetter(c, "secret")
}

func TestKVEnvDiffer_Unchanged(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kvEnvDiffResponse(data)(w, r)
	}))
	defer srv.Close()

	left := makeEnvGetter(t, srv.URL)
	right := makeEnvGetter(t, srv.URL)
	d := NewKVEnvDiffer(left, right)

	res, err := d.Compare("secret", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "unchanged" {
		t.Errorf("expected unchanged, got %q", res.Status)
	}
}

func TestKVEnvDiffer_Modified(t *testing.T) {
	srvLeft := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kvEnvDiffResponse(map[string]interface{}{"key": "old"})(w, r)
	}))
	defer srvLeft.Close()

	srvRight := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kvEnvDiffResponse(map[string]interface{}{"key": "new"})(w, r)
	}))
	defer srvRight.Close()

	d := NewKVEnvDiffer(makeEnvGetter(t, srvLeft.URL), makeEnvGetter(t, srvRight.URL))
	res, err := d.Compare("secret", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "modified" {
		t.Errorf("expected modified, got %q", res.Status)
	}
}

func TestKVEnvDiffer_SortedKeys(t *testing.T) {
	res := &KVEnvDiffResult{
		Left:  map[string]string{"z": "1", "a": "2"},
		Right: map[string]string{"m": "3"},
	}
	keys := res.SortedKeys()
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("unexpected key order: %v", keys)
	}
}

func TestNewKVEnvDiffer_NotNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	g := makeEnvGetter(t, srv.URL)
	d := NewKVEnvDiffer(g, g)
	if d == nil {
		t.Fatal("expected non-nil KVEnvDiffer")
	}
}
