package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func auditDeviceResponse() map[string]interface{} {
	return map[string]interface{}{
		"file/": map[string]interface{}{
			"type":        "file",
			"description": "file audit log",
			"options":     map[string]string{"file_path": "/var/log/vault.log"},
		},
		"syslog/": map[string]interface{}{
			"type":        "syslog",
			"description": "syslog audit",
			"options":     map[string]string{},
		},
	}
}

func TestListAuditDevices_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/audit" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(auditDeviceResponse())
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "test-token", HTTP: ts.Client()}
	lister := NewAuditDeviceLister(c)
	dev()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestListAuditDevices_Forbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "bad-token", HTTP: ts.Client()}
	lister := NewAuditDeviceLister(c)
	_, err := lister.List()
	if err == nil || err.Error() != "permission denied" {
		t.Errorf("expected permission denied, got %v", err)
	}
}

func TestListAuditDevices_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{Address: ts.URL, Token: "tok", HTTP: ts.Client()}
	lister := NewAuditDeviceLister(c)
	_, err := lister.List()
	if err == nil {
		t.Error("expected error for 500 status")
	}
}
